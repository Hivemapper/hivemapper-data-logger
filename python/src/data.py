import pandas as pd
import sqlite3
from sqlite3 import Error
import numpy as np


def create_connection(db_file: str):
    conn = None
    try:
        conn = sqlite3.connect(db_file, detect_types=sqlite3.PARSE_DECLTYPES | sqlite3.PARSE_COLNAMES)
    except Error as e:
        print(e)

    return conn


def create_corrected_gps_data_table(conn):
    create_table_sql = """
        CREATE TABLE IF NOT EXISTS corrected_gps_data(
            id INTEGER NOT NULL PRIMARY KEY,
            lat REAL NOT NULL,
            long REAL NOT NULL
        );
    """
    try:
        cursor = conn.cursor()
        cursor.execute(create_table_sql)
    except Error as e:
        print("failed to create corrected gps data table", e)
        raise e


def save_corrected_gps_data(corrected_gps_data, conn):
    return None


# Accelerometer Data
def fetch_accelerometer_data(conn):
    print('Fetching accelerometer data...')
    select_query = "SELECT CAST(STRFTIME('%s', imu_time) AS DECIMAL) AS imu_time_ms, imu_acc_x, imu_acc_y, imu_acc_z FROM imu_raw;"
    accel_columns = ["time", "x", "y", "z"]
    accel_df = pd.read_sql_query(select_query, conn)
    accel_df = accel_df.rename(columns={'imu_time_ms': 'time', 'imu_acc_x': 'x', 'imu_acc_y': 'y', 'imu_acc_z': 'z'})
    dtypes = {k: float for k in accel_columns[:4]}
    accel_df = accel_df.astype(dtypes)

    # Check if data is doubled up. This is happening in the exiftool extraction step. I am not sure how to remedy this
    # halfway_A = (int(np.shape(accel_df)[0] / 2))
    # print("Before: ", np.shape(accel_df))
    # if accel_df["time"][0] == accel_df["time"][halfway_A]:
    #     print("Accelerometer data doubled up. Removing half")
    #     accel_df = accel_df[:][0:halfway_A]
    #     print("After: ", np.shape(accel_df))

    # median2 = accel_df["y"].median()
    # median3 = accel_df["z"].median()
    # accel_df = accel_df.sub([0, 0, median2, median3], axis='columns')

    return accel_df


# GPS Data
def fetch_gps_data(conn):
    print("Fetching gps data...")
    select_query = "SELECT gnss_system_time, gnss_latitude, gnss_longitude, gnss_altitude, gnss_speed, gnss_heading FROM imu_raw;"
    gps_columns = ["time", "lat", "lon", "alt", "abs_vel", "heading"]
    gps_df = pd.read_sql_query(select_query, conn)
    gps_df = gps_df.rename(columns={'gnss_system_time': 'time', 'gnss_latitude': 'lat', 'gnss_longitude': 'lon', 'gnss_altitude': 'alt', 'gnss_speed': 'abs_vel', 'gnss_heading': 'heading'})
    gps_df["time"] = pd.to_datetime(gps_df["time"], errors="coerce", format='%Y:%m:%d %H:%M:%S')
    dtypes = {k: float for k in gps_columns[1:6]}
    gps_df = gps_df.astype(dtypes)

    # Check if data is doubled up. This is happening in the exiftool extraction step. I am not sure how to remedy this
    # print("Before: ", np.shape(gps_df))
    # halfway_G = (int(np.shape(gps_df)[0] / 2))
    # if gps_df["time"][0] == gps_df["time"][halfway_G]:
    #     print("GPS data doubled up. Removing half")
    #     gps_df = gps_df[:][0:halfway_G]
    #     print("After: ", np.shape(gps_df))

    return gps_df
