import numpy as np
import sys
from matplotlib import pyplot as plt
import utm

from data import fetch_gps_data, create_connection, fetch_accelerometer_data, save_corrected_gps_data, \
    create_corrected_gps_data_table
from kalman_filtering_gps_acceleration import make_H, make_Uin, make_F, make_G, make_Q, predict_state, \
    predict_covariance, make_R, make_K, update_state, update_covariance, latlon_to_utm, utm_to_latlon


def main(db_file):
    print("Initializing kalman filtering...")
    # Initialize kalman filtering stuff
    running_kalman = True
    gps_count = 0  # sensorS is for the time total time in seconds base from the first time we have (not 0 as its nan)
    imu_count = 0  # sensorC is the time itself
    imu_active = True

    corrected_gps_long = []
    corrected_gps_lat = []
    H = make_H()

    # Initialize sensor stuff
    imu_time_old = 0.0
    uncertainty_GPS = 4.0  # +/- meters (pretty standard for GPS sensors)
    guess_var = 100.0  # 100.0 is a totally random guess

    conn = create_connection(db_file)
    gps_df = fetch_gps_data(conn)
    imu_df = fetch_accelerometer_data(conn)

    # first item in the gps_dataframe is nan, we take the [1] as the _first_ item
    times = [element.total_seconds() for element in (gps_df["time"] - gps_df["time"][1])]
    imu_times = [element for element in list(imu_df["time"])]

    # Convert lat/lon to ECEF (N,E)
    utmN, utmE = latlon_to_utm(gps_df["lat"].values.tolist(), gps_df["lon"].values.tolist())
    heading = (gps_df["heading"]).values.tolist()
    gps_utm_data = [utmE, utmN, heading]

    # This data is in camera-based coordinate system
    imu_accel_data = [list(imu_df["x"]), list(imu_df["y"]), list(imu_df["z"])]
    #################################################################
    input_gps_location = np.zeros((4, 1))

    # Initialize position
    input_gps_location[0] = gps_utm_data[0][0]  # Easting
    input_gps_location[2] = gps_utm_data[1][0]  # Northing

    # Initialize velocity
    vel = (gps_df["abs_vel"]).values.tolist()
    a = make_Uin(0.0, vel[1], heading[1])
    input_gps_location[1] = a[0]
    input_gps_location[3] = a[1]
    P_in = np.identity(4) * guess_var

    last_gps_time = -1

    while running_kalman:
        # Get timestamps
        imu_time = imu_times[imu_count]
        gps_time = times[imu_count]

        # TODO How do we define the heading for an initial run?
        heading = gps_utm_data[2][gps_count]  # In degrees from North
        aF = imu_accel_data[2][imu_count]
        aR = imu_accel_data[1][imu_count]
        U_in = make_Uin(aR, aF, heading)  # Only want the 2 planar components of acceleration

        # Get dt -> time difference between imu and gps
        dt = imu_time - imu_time_old

        if imu_time_old == -1:
            imu_count = imu_count + 1
            imu_time_old = imu_time
            continue

        # Save state values for output
        corrected_gps_long.append(input_gps_location[0])
        corrected_gps_lat.append(input_gps_location[2])

        ###########################################
        # PREDICT (dynamic / state transition)
        F = make_F(dt)
        G = make_G(dt)
        Q = make_Q(G)
        predicted_gps_position = predict_state(input_gps_location, U_in, F, G)
        P_tmp = predict_covariance(P_in, F, Q)
        ###########################################

        if last_gps_time == gps_time:
            X = predicted_gps_position
            P = P_tmp
        else:
            R = make_R(uncertainty_GPS)
            K = make_K(H, P_tmp, R)
            measurement = [[gps_utm_data[0][gps_count]], [gps_utm_data[1][gps_count]]]
            X = update_state(predicted_gps_position, K, H, measurement)
            P = update_covariance(P_tmp, K, H, R)

        # reset time
        last_gps_time = gps_time

        # Update position
        input_gps_location = X
        P_in = P
        # Update the sensor count for the sensor that was used

        # taking next time and reset time sensor
        imu_count = imu_count + 1
        imu_time_old = imu_time

        if imu_count == len(imu_times) - 1:
            imu_active = False

        if not imu_active:
            running_kalman = False

    # Plot the output results

    # original
    plt.scatter(gps_utm_data[1], gps_utm_data[0])  # (lat, lon))
    plt.title("Original GPS Data")
    plt.xlabel("latitude")
    plt.ylabel("longitude")
    plt.show()

    # corrected
    plt.scatter(corrected_gps_lat[0:], corrected_gps_long[0:])  # (lat, lon))
    plt.title("Kalman Filtered GPS Data")
    plt.xlabel("latitude")
    plt.ylabel("longitude")
    plt.show()

    lat_long_degrees = utm_to_latlon(corrected_gps_long, corrected_gps_lat)
    create_corrected_gps_data_table(conn)
    save_corrected_gps_data(conn, lat_long_degrees)


if __name__ == '__main__':
    if len(sys.argv) == 2:
        main(sys.argv[1])
        print("kalman filtering done")
    else:
        print("no database file provided")
