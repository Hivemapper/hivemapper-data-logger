import numpy as np
import sys
from matplotlib import pyplot as plt

from data import fetch_gps_data, create_connection, fetch_accelerometer_data
from kalman_filtering_gps_acceleration import make_H, make_Uin, make_F, make_G, make_Q, predict_state, \
    predict_covariance, make_R, make_K, update_state, update_covariance, find_heading, latlon_to_utm


def main(db_file):
    print("Initializing kalman filtering...")
    # Initialize kalman filtering stuff
    running_kalman = True
    gps_count = 0  # sensorS is for the time total time in seconds base from the first time we have (not 0 as its nan)
    imu_count = 0  # sensorC is the time itself
    gps_active = True
    imu_active = True
    use_gps_time = False
    use_imu_time = False
    X_pos = []
    Y_pos = []
    H = make_H()

    # Initialize sensor stuff
    gps_time = 0.0
    imu_time = 0.0
    time_sensor_old = 0.0
    uncertainty_GPS = 4.0  # +/- meters (pretty standard for GPS sensors)
    guess_var = 100.0  # 100.0 is a totally random guess
    dt_diff_thresh = 0.09  # [sec]

    conn = create_connection(db_file)
    gps_df = fetch_gps_data(conn)
    imu_df = fetch_accelerometer_data(conn)

    # first item in the gps_dataframe is nan, we take the [1] as the _first_ item
    gps_times = [element.total_seconds() for element in (gps_df["time"] - gps_df["time"][1])]
    imu_times = [element for element in list(imu_df["time"])]

    # Convert lat/lon to ECEF (N,E)
    utmN, utmE = latlon_to_utm(gps_df["lat"].values.tolist(), gps_df["lon"].values.tolist())
    heading = (gps_df["heading"]).values.tolist()
    gps_utm_data = [utmE, utmN, heading]

    # This data is in camera-based coordinate system
    imu_accel_data = [list(imu_df["x"]), list(imu_df["y"]), list(imu_df["z"])]
    #################################################################
    X_in = np.zeros((4, 1))

    # Initialize position
    X_in[0] = gps_utm_data[0][0]  # Easting
    X_in[2] = gps_utm_data[1][0]  # Northing

    # Initialize velocity
    vel = (gps_df["abs_vel"]).values.tolist()
    a = make_Uin(0.0, vel[1], heading[1])
    X_in[1] = a[0]
    X_in[3] = a[1]
    P_in = np.identity(4) * guess_var

    while running_kalman:
        # Get timestamps
        if gps_active:
            gps_time = gps_times[gps_count]
        else:
            gps_time = 999999999999999.9
        if imu_active:
            imu_time = imu_times[imu_count]
        else:
            imu_time = 999999999999999.9
        # Choose which operation to perform (prediction or prediction+update)
        if abs(gps_time - imu_time) < dt_diff_thresh:  # "Exact" match! Let's do prediction AND update
            # fixme we never really go here
            # print("Exact match")
            use_gps_time = True
            use_imu_time = True
            # TODO How do we define the heading for an initial run?
            heading = gps_utm_data[2][gps_count]  # In degrees from North
            aF = imu_accel_data[2][imu_count]
            aR = imu_accel_data[1][imu_count]
            U_in = make_Uin(aR, aF, heading)  # Only want the 2 planar components of acceleration
        else:  # Do prediction ONLY
            # print("Prediction only")
            use_imu_time = True
            aF = imu_accel_data[2][imu_count]
            aR = imu_accel_data[1][imu_count]
            heading = gps_utm_data[2][gps_count]  # In degrees from North
            U_in = make_Uin(aR, aF, heading)  # Only want the 2 planar components of acceleration
        # Get dt
        dtC = imu_time - time_sensor_old

        dtS = gps_time - time_sensor_old
        dt = min(dtC, dtS)  # TODO: Is this really what we want to do?
        # Save state values for output
        X_pos.append(X_in[0])
        Y_pos.append(X_in[2])

        ###########################################
        # PREDICT (dynamic / state transition)
        F = make_F(dt)
        G = make_G(dt)
        Q = make_Q(G)
        X_tmp = predict_state(X_in, U_in, F, G)
        P_tmp = predict_covariance(P_in, F, Q)
        ###########################################

        if use_gps_time:
            # print("use GPS")
            # UPDATE with GPS measurement
            R = make_R(uncertainty_GPS)
            K = make_K(H, P_tmp, R)
            measurement = [[gps_utm_data[0][gps_count]], [gps_utm_data[1][gps_count]]]
            X = update_state(X_tmp, K, H, measurement)
            P = update_covariance(P_tmp, K, H, R)
        else:  # imu_time
            # print("dont use GPS")
            X = X_tmp
            P = P_tmp
        # Update heading based on previous X
        dE = X[0] - X_in[0]
        dN = X[2] - X_in[2]
        if not (dE == 0 and dN == 0):  # We don't update if no change in position
            heading = find_heading(dE, dN)
        # Update position
        X_in = X
        P_in = P
        # Update the sensor count for the sensor that was used
        if use_gps_time:
            if gps_count != len(gps_times) - 1:
                gps_count = gps_count + 1
            time_sensor_old = gps_time
            use_gps_time = False
        if use_imu_time:
            if imu_count != len(imu_times) - 1:
                imu_count = imu_count + 1
            # fixme: when this is commented, we loop infinity
            # else:
            #     use_gps_time = True
            time_sensor_old = imu_time  # TODO: What about the case where use_gps_time=True and use_sensorC=True????
            use_imu_time = False
        # Stop using sensors when finished
        if gps_count == len(gps_times) - 1:
            gps_active = False
        if imu_count == len(imu_times) - 1:
            imu_active = False
        # print(gps_active,imu_active,gps_count,len(gps_times))
        # End loop if necessary
        # if (not gps_active) and (not imu_active):
        if (not imu_active):
            running_kalman = False

    # Plot the output results

    # original
    plt.scatter(gps_utm_data[1], gps_utm_data[0])
    plt.show()

    # corrected
    plt.scatter(Y_pos[0:], X_pos[0:])
    plt.show()

    # print(Y_pos[1:])
    # print(X_pos[1:])
    # print(gps_utm_data[2])
    # ################################################################################
    # # Display ACCEL
    # import altair as alt
    # alt.Chart(imu_df).mark_point().encode(
    #   x='time',
    #   y='a2',
    # ).interactive()


if __name__ == '__main__':
    if len(sys.argv) == 2:
        main(sys.argv[1])
        print("kalman filtering done")
    else:
        print("no database file provided")
