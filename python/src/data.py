import pandas as pd
import sqlite3
from sqlite3 import Error


def create_connection(db_file: str):
    conn = None
    try:
        conn = sqlite3.connect(db_file, detect_types=sqlite3.PARSE_DECLTYPES | sqlite3.PARSE_COLNAMES)
    except Error as e:
        print(e)

    return conn


# Accelerometer Data
def fetch_accelerometer_data(conn):
    accel_columns = ["time", "x", "y", "z"]
    accel_df = pd.read_sql_query("select imu_time, imu_acc_x, imu_acc_y, imu_acc_z from imu_raw;", conn)
    accel_df = accel_df.rename(columns={'imu_time': 'time', 'imu_acc_x': 'x', 'imu_acc_y': 'y', 'imu_acc_z': 'z'})
    dtypes = {k: float for k in accel_columns[1:4]}
    accel_df["time"] = pd.to_datetime(accel_df["time"], errors="coerce", format='%Y:%m:%d %H:%M:%S')
    accel_df = accel_df.astype(dtypes)
    return accel_df


# GPS Data
def fetch_gps_data(conn):
    select_query = 'select gnss_system_time as gnss_system_date, gnss_latitude, gnss_longitude, gnss_altitude, gnss_speed, gnss_heading from imu_raw;'
    cursor = conn.cursor()
    cursor.execute(select_query)
    records = cursor.fetchall()
    gps_columns = ["time", "lat", "lon", "alt", "abs_vel", "heading"]
    # gps_df = pd.read_sql_query(select_query, conn)
    gps_df = pd.DataFrame(records)
    gps_df = gps_df.rename(columns={'gnss_system_time': 'time', 'gnss_latitude': 'lat', 'gnss_longitude': 'lon', 'gnss_altitude': 'alt', 'gnss_speed': 'abs_vel', 'gnss_heading': 'heading'})
    # gps_df["time"] = gps_df["time"].str.split(".00Z", n=1, expand=True)

    gps_df["time"] = pd.to_datetime(gps_df["time"], errors="coerce", format='%Y:%m:%d %H:%M:%S')

    dtypes = {k: float for k in gps_columns[1:6]}
    gps_df = gps_df.astype(dtypes)
    return gps_df
