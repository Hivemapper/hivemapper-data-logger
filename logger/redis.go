package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	sensordata "github.com/Hivemapper/hivemapper-data-logger/proto-out"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/go-redis/redis/v8"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type MagnetometerRedisWrapper struct {
	System_time time.Time `json:"system_time"`
	Mag_x       float64   `json:"mag_x"`
	Mag_y       float64   `json:"mag_y"`
	Mag_z       float64   `json:"mag_z"`
}

type ImuRedisWrapper struct {
	System_time time.Time `json:"system_time"`
	Accel       *Accel    `json:"accel"`
	Gyro        *Gyro     `json:"gyro"`
	Temp        float64   `json:"temp"`
	Time        time.Time `json:"time"`
	Fsync       *Fsync    `json:"fsync"`
}

func NewImuRedisWrapper(system_time time.Time, temperature iim42652.Temperature, acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, fsync *iim42652.Fsync) *ImuRedisWrapper {
	return &ImuRedisWrapper{
		System_time: system_time,
		Accel:       NewAccel(acceleration.X, acceleration.Y, acceleration.Z),
		Gyro:        NewGyro(angularRate.X, angularRate.Y, angularRate.Z),
		Time:        acceleration.Time,
		Temp:        *temperature,
		Fsync:       NewFsync(fsync.TimeDelta, fsync.FsyncInt),
	}
}

func NewMagnetometerRedisWrapper(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) *MagnetometerRedisWrapper {
	return &MagnetometerRedisWrapper{
		System_time: system_time,
		Mag_x:       mag_x,
		Mag_y:       mag_y,
		Mag_z:       mag_z,
	}
}

type Redis struct {
	DB                 *redis.Client
	ctx                context.Context
	maxImuEntries      int
	maxMagEntries      int
	maxGnssEntries     int
	maxGnssAuthEntries int
	logProtoText       bool
}

func NewRedis(maxImuEntries int, maxMagEntries int, maxGnssEntries int, maxGnssAuthEntries int, logProtoText bool) *Redis {
	return &Redis{
		maxImuEntries:      maxImuEntries,
		maxMagEntries:      maxMagEntries,
		maxGnssEntries:     maxGnssEntries,
		maxGnssAuthEntries: maxGnssAuthEntries,
		logProtoText:       logProtoText,
	}
}

var (
	prevItowMs uint32
)

func (s *Redis) Init() error {
	fmt.Println("Initializing Redis logger")
	s.DB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Use your Redis server address
	})
	fmt.Println("Getting context")
	s.ctx = context.Background()

	// Test the connection with a PING command
	pong, err := s.DB.Ping(s.ctx).Result()
	if err != nil {
		fmt.Printf("Could not connect to Redis: %v\n", err)
		return fmt.Errorf("ping pong failed")
	}
	fmt.Println("Redis connected:", pong)
	fmt.Println("Redis logger initialized")
	return nil
}

func (s *Redis) LogImuData(imudata ImuRedisWrapper) error {
	// create imu proto
	newdata := sensordata.ImuData{
		SystemTime: imudata.System_time.String(),
		Accelerometer: &sensordata.ImuData_AccelerometerData{
			X: imudata.Accel.X,
			Y: imudata.Accel.Y,
			Z: imudata.Accel.Z,
		},
		Gyroscope: &sensordata.ImuData_GyroscopeData{
			X: imudata.Gyro.X,
			Y: imudata.Gyro.Y,
			Z: imudata.Gyro.Z,
		},
		Temperature: imudata.Temp,
		Time:        imudata.Time.String(),
		Fsync: &sensordata.ImuData_FsyncData{
			FsyncInt:  imudata.Fsync.FsyncInt,
			TimeDelta: int32(imudata.Fsync.TimeDelta),
		},
	}
	// serialize the data
	protodata, err := s.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "imu_data", protodata).Err(); err != nil {
		return err
	}

	if err := s.DB.LTrim(s.ctx, "imu_data", 0, int64(s.maxImuEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogMagnetometerData(magdata MagnetometerRedisWrapper) error {
	// create magnetometer proto
	newdata := sensordata.MagnetometerData{
		SystemTime: magdata.System_time.String(),
		X:          magdata.Mag_x,
		Y:          magdata.Mag_y,
		Z:          magdata.Mag_z,
	}
	// serialize the data
	protodata, err := s.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "magnetometer_data", protodata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "magnetometer_data", 0, int64(s.maxMagEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssAuthData(gnssAuthData neom9n.Data) error {
	// Create gnss auth proto
	newdata := sensordata.GnssData{
		SystemTime: gnssAuthData.SystemTime.String(),
		SecEcsign: &sensordata.GnssData_UbxSecEcsign{
			Version:        uint32(gnssAuthData.SecEcsign.Version),
			MsgNum:         uint32(gnssAuthData.SecEcsign.MsgNum),
			FinalHash:      gnssAuthData.SecEcsign.FinalHash[:],
			SessionId:      gnssAuthData.SecEcsign.SessionId[:],
			EcdsaSignature: gnssAuthData.SecEcsign.EcdsaSignature[:],
		},
		SecEcsignBuffer: gnssAuthData.SecEcsignBuffer,
	}
	protodata, err := s.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_auth_data", protodata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_auth_data", 0, int64(s.maxGnssAuthEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) Marshal(message proto.Message) ([]byte, error) {
	var data []byte
	var err error

	if s.logProtoText {
		data, err = prototext.Marshal(message)
	} else {
		data, err = proto.Marshal(message)
	}

	return data, err
}

func (s *Redis) HandleUbxMessage(msg interface{}) error {
	systemTime := time.Now().UTC()
	var protodata []byte = nil
	var err error
	var redisKey string = "INVALID"

	if s.DB == nil {
		fmt.Println("Redis DB is nil")
		return nil
	}

	switch m := msg.(type) {
	case *ubx.NavPvt:
		redisKey = "NavPvt"
		// serialize as proto
		protomessage := sensordata.NavPvt{
			SystemTime:   systemTime.String(),
			ItowMs:       m.ITOW_ms,
			YearY:        uint32(m.Year_y),
			MonthMonth:   uint32(m.Month_month),
			DayD:         uint32(m.Day_d),
			HourH:        uint32(m.Hour_h),
			MinMin:       uint32(m.Min_min),
			SecS:         uint32(m.Sec_s),
			Valid:        uint32(m.Valid),
			TAccNs:       uint32(m.TAcc_ns),
			NanoNs:       uint32(m.Nano_ns),
			FixType:      uint32(m.FixType),
			Flags:        uint32(m.Flags),
			Flags2:       uint32(m.Flags2),
			NumSv:        uint32(m.NumSV),
			LonDege7:     int32(m.Lon_dege7),
			LatDege7:     int32(m.Lat_dege7),
			HeightMm:     int32(m.Height_mm),
			HmslMm:       int32(m.HMSL_mm),
			HAccMm:       uint32(m.HAcc_mm),
			VAccMm:       uint32(m.VAcc_mm),
			VelNMmS:      int32(m.VelN_mm_s),
			VelEMmS:      int32(m.VelE_mm_s),
			VelDMmS:      int32(m.VelD_mm_s),
			GSpeedMmS:    int32(m.GSpeed_mm_s),
			HeadMotDege5: int32(m.HeadMot_dege5),
			SAccMmS:      uint32(m.SAcc_mm_s),
			HeadAccDege5: int32(m.HeadAcc_dege5),
			Pdop:         uint32(m.PDOP),
			Flags3:       uint32(m.Flags3),
			HeadVehDege5: int32(m.HeadVeh_dege5),
			MagDecDege2:  int32(m.MagDec_dege2),
			MagAccDege2:  uint32(m.MagAcc_dege2),
		}

		if prevItowMs != 0 && m.ITOW_ms-prevItowMs > 250 {
			fmt.Println("[WARNING] NavPvt drop of", m.ITOW_ms-prevItowMs, "ms (", prevItowMs, ",", m.ITOW_ms, ")")
		}
		prevItowMs = m.ITOW_ms

		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavDop:
		redisKey = "NavDop"
		// serialize as proto
		protomessage := sensordata.NavDop{
			SystemTime: systemTime.String(),
			ItowMs:     m.ITOW_ms,
			Gdop:       uint32(m.GDOP),
			Pdop:       uint32(m.PDOP),
			Tdop:       uint32(m.TDOP),
			Vdop:       uint32(m.VDOP),
			Hdop:       uint32(m.HDOP),
			Ndop:       uint32(m.NDOP),
			Edop:       uint32(m.EDOP),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavCov:
		redisKey = "NavCov"
		// serialize as proto
		protomessage := sensordata.NavCov{
			ItowMs:      m.ITOW_ms,
			Version:     uint32(m.Version),
			PosCovValid: uint32(m.PosCovValid),
			VelCovValid: uint32(m.VelCovValid),
			PosCovNN:    float64(m.PosCovNN_m2),
			PosCovNE:    float64(m.PosCovNE_m2),
			PosCovND:    float64(m.PosCovND_m2),
			PosCovEE:    float64(m.PosCovEE_m2),
			PosCovED:    float64(m.PosCovED_m2),
			PosCovDD:    float64(m.PosCovDD_m2),
			VelCovNN:    float64(m.VelCovNN_m2_s2),
			VelCovNE:    float64(m.VelCovNE_m2_s2),
			VelCovND:    float64(m.VelCovND_m2_s2),
			VelCovEE:    float64(m.VelCovEE_m2_s2),
			VelCovED:    float64(m.VelCovED_m2_s2),
			VelCovDD:    float64(m.VelCovDD_m2_s2),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavPosecef:
		redisKey = "NavPosecef"
		// serialize as proto
		protomessage := sensordata.NavPosecef{
			ItowMs:  m.ITOW_ms,
			EcefXCm: int32(m.EcefX_cm),
			EcefYCm: int32(m.EcefY_cm),
			EcefZCm: int32(m.EcefZ_cm),
			PAccCm:  uint32(m.PAcc_cm),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavTimegps:
		redisKey = "NavTimegps"
		// serialize as proto
		protomessage := sensordata.NavTimegps{
			ItowMs: uint32(m.ITOW_ms),
			FtowNs: int32(m.FTOW_ns),
			Week:   int32(m.Week),
			LeapS:  int32(m.LeapS_s),
			Valid:  uint32(m.Valid),
			TAccNs: uint32(m.TAcc_ns),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavVelecef:
		redisKey = "NavVelecef"
		// serialize as proto
		protomessage := sensordata.NavVelecef{
			ItowMs:    uint32(m.ITOW_ms),
			EcefVxCmS: int32(m.EcefVX_cm_s),
			EcefVyCmS: int32(m.EcefVY_cm_s),
			EcefVzCmS: int32(m.EcefVZ_cm_s),
			SAccCmS:   uint32(m.SAcc_cm_s),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavStatus:
		redisKey = "NavStatus"
		// serialize as proto
		protomessage := sensordata.NavStatus{
			ItowMs:  uint32(m.ITOW_ms),
			GpsFix:  uint32(m.GpsFix),
			Flags:   uint32(m.Flags),
			FixStat: uint32(m.FixStat),
			Flags2:  uint32(m.Flags2),
			Ttff:    uint32(m.Ttff_ms),
			Msss:    uint32(m.Msss_ms),
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavSat:
		redisKey = "NavSat"
		protomessage := sensordata.NavSat{
			SystemTime: systemTime.String(),
			ItowMs:     m.ITOW_ms,
			Version:    uint32(m.Version),
			NumSvs:     uint32(m.NumSvs),
		}
		protomessage.Svs = make([]*sensordata.NavSat_Svs, len(m.Svs))
		for i, sv := range m.Svs {
			protomessage.Svs[i] = &sensordata.NavSat_Svs{
				GnssId:   uint32(sv.GnssId),
				SvId:     uint32(sv.SvId),
				CnoDbhz:  uint32(sv.Cno_dbhz),
				ElevDeg:  int32(sv.Elev_deg),
				AzimDeg:  int32(sv.Azim_deg),
				PrResMe1: int32(sv.PrRes_me1),
				Flags:    uint32(sv.Flags),
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.NavSig:
		redisKey = "NavSig"
		protomessage := sensordata.NavSig{
			SystemTime: systemTime.String(),
			ItowMs:     m.ITOW_ms,
			Version:    uint32(m.Version),
			NumSigs:    uint32(m.NumSigs),
		}
		protomessage.Sigs = make([]*sensordata.NavSig_Sigs, len(m.Sigs))
		for i, sig := range m.Sigs {
			protomessage.Sigs[i] = &sensordata.NavSig_Sigs{
				GnssId:     uint32(sig.GnssId),
				SvId:       uint32(sig.SvId),
				SigId:      uint32(sig.SigId),
				FreqId:     uint32(sig.FreqId),
				PrResMe1:   int32(sig.PrRes_me1),
				CnoDbhz:    uint32(sig.Cno_dbhz),
				QualityInd: uint32(sig.QualityInd),
				CorrSource: uint32(sig.CorrSource),
				IonoModel:  uint32(sig.IonoModel),
				SigFlags:   uint32(sig.SigFlags),
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.MonRf:
		redisKey = "MonRf"
		protomessage := sensordata.MonRf{
			SystemTime: systemTime.String(),
			Version:    uint32(m.Version),
			NBlock:     uint32(m.NBlock),
		}
		protomessage.RfBlocks = make([]*sensordata.MonRf_RFBlock, m.NBlock)
		for i, block := range m.RFBlocks {
			protomessage.RfBlocks[i] = &sensordata.MonRf_RFBlock{
				BlockId:    uint32(block.BlockId),
				Flags:      uint32(block.Flags),
				AntStatus:  uint32(block.AntStatus),
				AntPower:   uint32(block.AntPower),
				PostStatus: uint32(block.PostStatus),
				NoisePerMs: uint32(block.NoisePerMS),
				AgcCnt:     uint32(block.AgcCnt),
				JamInd:     int32(block.JamInd),
				OfsI:       int32(block.OfsI),
				MagI:       uint32(block.MagI),
				OfsQ:       int32(block.OfsQ),
				MagQ:       uint32(block.MagQ),
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.RxmMeasx:
		redisKey = "RxmMeasx"
		protomessage := sensordata.RxmMeasx{
			SystemTime:     systemTime.String(),
			Version:        uint32(m.Version),
			GpsTowMs:       m.GpsTOW_ms,
			GloTowMs:       m.GloTOW_ms,
			BdsTowMs:       m.BdsTOW_ms,
			QzssTowMs:      m.QzssTOW_ms,
			GpsTowAccMsl4:  uint32(m.GpsTOWacc_msl4),
			GloTowAccMsl4:  uint32(m.GloTOWacc_msl4),
			BdsTowAccMsl4:  uint32(m.BdsTOWacc_msl4),
			QzssTowAccMsl4: uint32(m.QzssTOWacc_msl4),
			NumSv:          uint32(m.NumSV),
			Flags:          uint32(m.Flags),
		}
		protomessage.Sv = make([]*sensordata.RxmMeasx_RxmMeasxSVType, len(m.SV))
		for i, sv := range m.SV {
			protomessage.Sv[i] = &sensordata.RxmMeasx_RxmMeasxSVType{
				GnssId:          uint32(sv.GnssId),
				SvId:            uint32(sv.SvId),
				CNo:             uint32(sv.CNo),
				MpathIndic:      uint32(sv.MpathIndic),
				DopplerMsMS:     int32(sv.DopplerMS_m_s),
				DopplerHzHz:     int32(sv.DopplerHz_hz),
				WholeChips:      uint32(sv.WholeChips),
				FracChips:       uint32(sv.FracChips),
				CodePhaseMsl_21: uint32(sv.CodePhase_msl21),
				IntCodePhaseMs:  uint32(sv.IntCodePhase_ms),
				PseuRangeRmsErr: uint32(sv.PseuRangeRMSErr),
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.RxmRawx:
		redisKey = "RxmRawx"
		protomessage := sensordata.RxmRawx{
			SystemTime: systemTime.String(),
			RcvTowS:    m.RcvTow_s,
			Week:       uint32(m.Week_weeks),
			LeapS:      uint32(m.LeapS_s),
			NumMeas:    uint32(m.NumMeas),
			RecStat:    uint32(m.RecStat),
			Version:    uint32(m.Version),
		}
		protomessage.Meas = make([]*sensordata.RxmRawx_RxmRawxMeasType, len(m.Meas))
		for i, meas := range m.Meas {
			protomessage.Meas[i] = &sensordata.RxmRawx_RxmRawxMeasType{
				PrMes:             meas.PrMes_m,
				CpMes:             meas.CpMes_cycles,
				DoMes:             float64(meas.DoMes_hz),
				GnssId:            uint32(meas.GnssId),
				SvId:              uint32(meas.SvId),
				SigId:             uint32(meas.SigId),
				FreqId:            uint32(meas.FreqId),
				LocktimeMs:        uint32(meas.Locktime_ms),
				CnoDbhz:           uint32(meas.Cno_dbhz),
				PrStdevM_1E2_2N:   uint32(meas.PrStdev_m),
				CpStdevCycles_4E3: uint32(meas.CpStdev_cycles),
				DoStdevHz_2E3_2N:  uint32(meas.DoStdev_hz),
				TrkStat:           uint32(meas.TrkStat),
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.RxmSfrbx:
		redisKey = "RxmSfrbx"
		protomessage := sensordata.RxmSfrbx{
			SystemTime: systemTime.String(),
			GnssId:     uint32(m.GnssId),
			SvId:       uint32(m.SvId),
			SigId:      uint32(m.Reserved1),
			FreqId:     uint32(m.FreqId),
			NumWords:   uint32(m.NumWords),
			Chn:        uint32(m.Chn),
			Version:    uint32(m.Version),
		}
		protomessage.WordBlock = make([]*sensordata.RxmSfrbx_WordBlock, len(m.Words))
		for i, word := range m.Words {
			// Note: this is a byte array, so we will just pass the bytes directly
			protomessage.WordBlock[i] = &sensordata.RxmSfrbx_WordBlock{
				Dwrd: uint32(word.Dwrd), // 32-bit word
			}
		}
		protodata, err = s.Marshal(&protomessage)
	case *ubx.TimTp:
		redisKey = "TimTp"
		protomessage := sensordata.TimTp{
			SystemTime: systemTime.String(),
			ItowMs:     uint32(m.TowMS_ms),
			ItowSubMs:  uint32(m.TowSubMS_msl32),
			QErrPs:     int32(m.QErr_ps),
			Week:       uint32(m.Week_weeks),
			Flags:      uint32(m.Flags),
			RefInfo:    uint32(m.RefInfo),
		}
		protodata, err = s.Marshal(&protomessage)
	}

	if protodata == nil {
		println("Redis key skipped:", redisKey)
		return nil
	}

	if err != nil {
		return err
	}

	// Push the proto data to the Redis list
	if err := s.DB.LPush(s.ctx, redisKey, protodata).Err(); err != nil {
		return err
	}

	// Trim the list to the max number of entries
	if err := s.DB.LTrim(s.ctx, redisKey, 0, int64(s.maxGnssEntries)).Err(); err != nil {
		return err
	}

	return nil
}
