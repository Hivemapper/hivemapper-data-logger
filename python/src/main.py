import numpy as np
import sys

from data import fetch_gps_data, create_connection, fetch_accelerometer_data
from kalman_filtering_gps_acceleration import make_H, make_Uin, make_F, make_G, make_Q, predict_state, \
    predict_covariance, make_R, make_K, update_state, update_covariance, find_heading, latlon_to_utm


def main(db_file):
    print("Initializing kalman filtering...")
    # Initialize kalman filtering stuff
    running_kalman = True
    sensorS_count = 1  # sensorS is for the time total time in seconds base from the first time we have (not 0 as its nan)
    sensorC_count = 0  # sensorC is the time itself
    sensorS_active = True
    sensorC_active = True
    use_sensorS = False
    use_sensorC = False
    X_pos = []
    Y_pos = []
    H = make_H()

    # Initialize sensor stuff
    time_sensorC = 0.0
    time_sensorS = 0.0
    time_sensor_old = 0.0
    uncertainty_GPS = 4.0  # +/- meters (pretty standard for GPS sensors)
    guess_var = 100.0  # 100.0 is a totally random guess
    dt_diff_thresh = 0.09  # [sec]

    conn = create_connection(db_file)
    gps_df = fetch_gps_data(conn)
    accel_df = fetch_accelerometer_data(conn)

    # first item in the gps_dataframe is nan, we take the [1] as the _first_ item
    sensorS_time = [element.total_seconds() for element in (gps_df["time"] - gps_df["time"][1])]
    sensorC_time = [element for element in list(accel_df["time"])]

    # Convert lat/lon to ECEF (N,E)
    utmN, utmE = latlon_to_utm(gps_df["lat"].values.tolist(), gps_df["lon"].values.tolist())
    heading = (gps_df["heading"]).values.tolist()
    sensorS_data = [utmE, utmN, heading]

    # This data is in camera-based coordinate system
    sensorC_data = [list(accel_df["x"]), list(accel_df["y"]), list(accel_df["z"])]
    #################################################################
    X_in = np.zeros((4, 1))

    # Initialize position
    X_in[0] = sensorS_data[0][0]  # Easting
    X_in[2] = sensorS_data[1][0]  # Northing

    # Initialize velocity
    vel = (gps_df["abs_vel"]).values.tolist()
    a = make_Uin(0.0, vel[1], heading[1])
    X_in[1] = a[0]
    X_in[3] = a[1]
    P_in = np.identity(4) * guess_var

    while running_kalman:
        # Get timestamps
        if sensorS_active:
            time_sensorS = sensorS_time[sensorS_count]
        else:
            time_sensorS = 999999999999999.9
        if sensorC_active:
            time_sensorC = sensorC_time[sensorC_count]
        else:
            time_sensorC = 999999999999999.9
        # Choose which operation to perform (prediction or prediction+update)
        # fixme: time_sensorC is a timestamp here and we are comparing it with time_sensorS which is seconds, should we compare with total_seconds()?
        if time_sensorS < time_sensorC:
            # print("Skip forward")
            # Just iterate forward. This step is based on the assumption that the Control (accelerometer) data is at a hig
            sensorS_count = sensorS_count + 1
            continue
        if abs(time_sensorS - time_sensorC) < dt_diff_thresh:  # "Exact" match! Let's do prediction AND update
            # print("Exact match")
            use_sensorS = True
            use_sensorC = True
            # TODO How do we define the heading for an initial run?
            heading = sensorS_data[2][sensorS_count]  # In degrees from North
            aF = sensorC_data[2][sensorC_count]
            aR = sensorC_data[1][sensorC_count]
            U_in = make_Uin(aR, aF, heading)  # Only want the 2 planar components of acceleration
        else:  # Do prediction ONLY
            # print("Prediction only")
            use_sensorC = True
            aF = sensorC_data[2][sensorC_count]
            aR = sensorC_data[1][sensorC_count]
            U_in = make_Uin(aR, aF, heading)  # Only want the 2 planar components of acceleration
        # Get dt
        dtC = time_sensorC - time_sensor_old

        dtS = time_sensorS - time_sensor_old
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
        if use_sensorS:
            # print("use GPS")
            # UPDATE with GPS measurement
            R = make_R(uncertainty_GPS)
            K = make_K(H, P_tmp, R)
            measurement = [[sensorS_data[0][sensorS_count]], [sensorS_data[1][sensorS_count]]]
            X = update_state(X_tmp, K, H, measurement)
            P = update_covariance(P_tmp, K, H, R)
        else:
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
        if use_sensorS:
            sensorS_count = sensorS_count + 1
            time_sensor_old = time_sensorS
            use_sensorS = False
        if use_sensorC:
            sensorC_count = sensorC_count + 1
            time_sensor_old = time_sensorC  # TODO: What about the case where use_sensorS=True and use_sensorC=True????
            use_sensorC = False
        # Stop using sensors when finished
        if sensorS_count == len(sensorS_time):
            sensorS_active = False
        if sensorC_count == len(sensorC_time):
            sensorC_active = False
        # print(sensorS_active,sensorC_active,sensorS_count,len(sensorS_time))
        # End loop if necessary
        if (not sensorS_active) and (not sensorC_active):
            running_kalman = False


if __name__ == '__main__':
    if len(sys.argv) == 2:
        main(sys.argv[1])
    print("no database file provided")
