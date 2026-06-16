#!/usr/bin/env python3
import json
import math
import sys
import argparse
from datetime import datetime, timedelta, timezone

# WGS-84 constants
WGS84_A = 6378137.0         # semi-major axis in meters
WGS84_E2 = 0.00669437999014 # first eccentricity squared

def latlon_to_ecef(lat_deg, lon_deg, alt_m):
    lat_rad = math.radians(lat_deg)
    lon_rad = math.radians(lon_deg)
    
    sin_lat = math.sin(lat_rad)
    cos_lat = math.cos(lat_rad)
    sin_lon = math.sin(lon_rad)
    cos_lon = math.cos(lon_rad)
    
    N = WGS84_A / math.sqrt(1.0 - WGS84_E2 * sin_lat * sin_lat)
    
    x = (N + alt_m) * cos_lat * cos_lon
    y = (N + alt_m) * cos_lat * sin_lon
    z = (N * (1.0 - WGS84_E2) + alt_m) * sin_lat
    
    return x, y, z

def ned_to_ecef_vel(lat_deg, lon_deg, vel_n, vel_e, vel_d):
    lat_rad = math.radians(lat_deg)
    lon_rad = math.radians(lon_deg)
    
    sin_lat = math.sin(lat_rad)
    cos_lat = math.cos(lat_rad)
    sin_lon = math.sin(lon_rad)
    cos_lon = math.cos(lon_rad)
    
    # NED to ECEF rotation matrix
    vx = -sin_lat * cos_lon * vel_n - sin_lon * vel_e - cos_lat * cos_lon * vel_d
    vy = -sin_lat * sin_lon * vel_n + cos_lon * vel_e - cos_lat * sin_lon * vel_d
    vz = cos_lat * vel_n - sin_lat * vel_d
    
    return vx, vy, vz

def generate_simulation(duration_sec, output_path, speed_kmh, start_lat, start_lon, heading_deg):
    # Calculate speed in meters/second
    speed_mps = speed_kmh / 3.6
    
    # Epoch rate: 4Hz (every 250ms)
    dt = 0.25
    steps = int(duration_sec / dt)
    
    # Start time
    now = datetime.now(timezone.utc)
    # Start GPS Time of Week (TOW) - let's pick a default
    start_itow_ms = 517405500
    start_uptime_ms = 25326299.0
    
    lat = start_lat
    lon = start_lon
    alt_m = 30.0  # Constant altitude (30 meters in SF)
    
    heading_rad = math.radians(heading_deg)
    
    with open(output_path, 'w') as f:
        for step in range(steps):
            t_offset = step * dt
            step_time = now + timedelta(seconds=t_offset)
            itow_ms = start_itow_ms + int(t_offset * 1000)
            uptime_ms = start_uptime_ms + (t_offset * 1000)
            
            # 1. Update position (Dead reckoning)
            ds = speed_mps * dt  # distance traveled in this step
            lat_rad = math.radians(lat)
            
            new_lat_rad = lat_rad + (ds * math.cos(heading_rad) / WGS84_A)
            new_lon_rad = math.radians(lon) + (ds * math.sin(heading_rad) / (WGS84_A * math.cos(lat_rad)))
            
            lat = math.degrees(new_lat_rad)
            lon = math.degrees(new_lon_rad)
            
            # Convert to ECEF (centimeters)
            x_m, y_m, z_m = latlon_to_ecef(lat, lon, alt_m)
            ecef_x_cm = int(x_m * 100)
            ecef_y_cm = int(y_m * 100)
            ecef_z_cm = int(z_m * 100)
            
            # Calculate velocities
            vel_n_m_s = speed_mps * math.cos(heading_rad)
            vel_e_m_s = speed_mps * math.sin(heading_rad)
            vel_d_m_s = 0.0
            
            g_speed_mm_s = int(round(speed_mps * 1000))
            vel_n_mm_s = int(round(vel_n_m_s * 1000))
            vel_e_mm_s = int(round(vel_e_m_s * 1000))
            vel_d_mm_s = int(round(vel_d_m_s * 1000))
            
            # Convert NED velocity to ECEF velocity
            ecef_vx_m_s, ecef_vy_m_s, ecef_vz_m_s = ned_to_ecef_vel(lat, lon, vel_n_m_s, vel_e_m_s, vel_d_m_s)
            ecef_vx_cm_s = int(round(ecef_vx_m_s * 100))
            ecef_vy_cm_s = int(round(ecef_vy_m_s * 100))
            ecef_vz_cm_s = int(round(ecef_vz_m_s * 100))
            
            head_mot_dege5 = int(round(heading_deg * 100000))
            head_veh_dege5 = int(round(heading_deg * 100000))
            
            lon_dege7 = int(lon * 1e7)
            lat_dege7 = int(lat * 1e7)
            
            # Format times
            sys_time_str = step_time.strftime('%Y-%m-%d %H:%M:%S') + f'.{step_time.microsecond * 1000:09d} +0000 UTC'
            rcv_tow_s = (itow_ms / 1000.0) + 0.004
            
            year = step_time.year
            month = step_time.month
            day = step_time.day
            hour = step_time.hour
            minute = step_time.minute
            second = step_time.second
            nano = (step_time.microsecond * 1000)
            
            # Write the epoch messages (sequence of 9 messages)
            
            # 1. RxmRawx
            rxm_rawx_data = (
                f'system_time:"{sys_time_str}" rcv_tow_s:{rcv_tow_s:.3f} week:2422 leap_s:18 num_meas:2 rec_stat:1 version:1 '
                f'meas:{{pr_mes:2.3631966949941665e+07 cp_mes:1.2418683236464132e+08 do_mes:-2038.1473388671875 sv_id:32 locktime_ms:8000 cno_dbhz:25 pr_stdev_m_1e2_2n:8 cp_stdev_cycles_4e3:10 do_stdev_hz_2e3_2n:8 trk_stat:7}} '
                f'meas:{{pr_mes:2.3793212786808778e+07 cp_mes:1.2503417617119361e+08 do_mes:339.675048828125 sv_id:2 cno_dbhz:23 pr_stdev_m_1e2_2n:10 cp_stdev_cycles_4e3:15 do_stdev_hz_2e3_2n:12 trk_stat:1}}'
            )
            f.write(json.dumps({"redisKey": "RxmRawx", "data": rxm_rawx_data}) + '\n')
            
            # 2. NavPvt
            nav_pvt_data = (
                f'system_time:"{sys_time_str}" itow_ms:{itow_ms} uptime_ms:{uptime_ms:.7e} '
                f'year_y:{year} month_month:{month} day_d:{day} hour_h:{hour} min_min:{minute} sec_s:{second} '
                f'valid:55 t_acc_ns:96186120 nano_ns:{nano} fix_type:3 flags:1 flags2:164 num_sv:10 lon_dege7:{lon_dege7} lat_dege7:{lat_dege7} '
                f'height_mm:{int(alt_m * 1000)} hmsl_mm:{int(alt_m * 1000)} h_acc_mm:2500 v_acc_mm:3800 '
                f'vel_n_mm_s:{vel_n_mm_s} vel_e_mm_s:{vel_e_mm_s} vel_d_mm_s:{vel_d_mm_s} g_speed_mm_s:{g_speed_mm_s} '
                f'head_mot_dege5:{head_mot_dege5} s_acc_mm_s:20008 head_acc_dege5:18000000 pdop:150 '
                f'head_veh_dege5:{head_veh_dege5}'
            )
            f.write(json.dumps({"redisKey": "NavPvt", "data": nav_pvt_data}) + '\n')
            
            # 3. NavPosecef
            nav_posecef_data = f'itow_ms:{itow_ms} ecef_x_cm:{ecef_x_cm} ecef_y_cm:{ecef_y_cm} ecef_z_cm:{ecef_z_cm} p_acc_cm:655312640'
            f.write(json.dumps({"redisKey": "NavPosecef", "data": nav_posecef_data}) + '\n')
            
            # 4. NavVelecef
            nav_velecef_data = f'itow_ms:{itow_ms} ecef_vx_cm_s:{ecef_vx_cm_s} ecef_vy_cm_s:{ecef_vy_cm_s} ecef_vz_cm_s:{ecef_vz_cm_s} s_acc_cm_s:2001'
            f.write(json.dumps({"redisKey": "NavVelecef", "data": nav_velecef_data}) + '\n')
            
            # 5. NavSig
            nav_sig_data = (
                f'system_time:"{sys_time_str}" itow_ms:{itow_ms} num_sigs:6 '
                f'sigs:{{sv_id:2 cno_dbhz:23 quality_ind:4 sig_flags:1}} '
                f'sigs:{{sv_id:2 sig_id:7 cno_dbhz:10 quality_ind:4 sig_flags:1}} '
                f'sigs:{{sv_id:6 cno_dbhz:24 quality_ind:3 sig_flags:1}} '
                f'sigs:{{sv_id:8 quality_ind:1 sig_flags:1}} '
                f'sigs:{{sv_id:32 cno_dbhz:25 quality_ind:7 sig_flags:1}} '
                f'sigs:{{sv_id:32 sig_id:7 quality_ind:1 sig_flags:1}}'
            )
            f.write(json.dumps({"redisKey": "NavSig", "data": nav_sig_data}) + '\n')
            
            # 6. NavStatus
            nav_status_data = f'itow_ms:{itow_ms} flags:76 flags2:8 msss:25336384'
            f.write(json.dumps({"redisKey": "NavStatus", "data": nav_status_data}) + '\n')
            
            # 7. NavDop
            nav_dop_data = (
                f'system_time:"{sys_time_str}" itow_ms:{itow_ms} gdop:180 pdop:150 tdop:100 vdop:120 hdop:90 ndop:80 edop:50'
            )
            f.write(json.dumps({"redisKey": "NavDop", "data": nav_dop_data}) + '\n')
            
            # 8. NavTimegps
            nav_timegps_data = f'itow_ms:{itow_ms} week:2422 leap_s:18 valid:7 t_acc_ns:96186096'
            f.write(json.dumps({"redisKey": "NavTimegps", "data": nav_timegps_data}) + '\n')
            
            # 9. NavCov
            nav_cov_data = (
                f'itow_ms:{itow_ms} pos_cov_valid:1 vel_cov_valid:1 '
                f'pos_cov_n_n:1.4314547183616e+13 pos_cov_n_e:-524288 pos_cov_n_d:4.718592e+06 '
                f'pos_cov_e_e:1.4314541940736e+13 pos_cov_e_d:-2.8573696e+07 pos_cov_d_d:1.4314380460032e+13 '
                f'vel_cov_n_n:400.31231689453125 vel_cov_n_e:0.0013885498046875 vel_cov_n_d:0.008087158203125 '
                f'vel_cov_e_e:400.3036193847656 vel_cov_e_d:-0.0516510009765625 vel_cov_d_d:400.0122375488281'
            )
            f.write(json.dumps({"redisKey": "NavCov", "data": nav_cov_data}) + '\n')
            
            # Output MonRf occasionally (every 10 seconds / 40 steps)
            if step % 40 == 0:
                mon_rf_data = (
                    f'system_time:"{sys_time_str}" n_block:2 '
                    f'rf_blocks:{{ant_status:1 ant_power:2 noise_per_ms:85 agc_cnt:1488 jam_ind:97 ofs_i:17 mag_i:255 ofs_q:13 mag_q:255}} '
                    f'rf_blocks:{{block_id:1 ant_status:1 ant_power:2 noise_per_ms:48 agc_cnt:2418 jam_ind:97 ofs_i:21 mag_i:255 ofs_q:21 mag_q:255}}'
                )
                f.write(json.dumps({"redisKey": "MonRf", "data": mon_rf_data}) + '\n')

    print(f"Simulation file successfully generated at: {output_path}")
    print(f"Duration: {duration_sec} seconds")
    print(f"Path traveled from: ({start_lat}, {start_lon}) to ({lat:.6f}, {lon:.6f})")

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description="Generate a simulated GNSS pbtxt/JSON replay log file.")
    parser.add_argument('--duration', type=float, default=60.0, help="Duration of simulation in seconds")
    parser.add_argument('--output', type=str, default="gnss_sim.txt", help="Output file path")
    parser.add_argument('--speed', type=float, default=50.0, help="Speed of the car in km/h")
    parser.add_argument('--lat', type=float, default=37.7749, help="Starting Latitude (default: SF)")
    parser.add_argument('--lon', type=float, default=-122.4194, help="Starting Longitude (default: SF)")
    parser.add_argument('--heading', type=float, default=90.0, help="Heading direction in degrees (0=North, 90=East, etc.)")
    
    args = parser.parse_args()
    generate_simulation(args.duration, args.output, args.speed, args.lat, args.lon, args.heading)
