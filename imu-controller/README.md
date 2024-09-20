# Imu Controller

This package contains the `ImuController` which is a controller that can be used to access acceleration and gyroscope data from an IMU sensor.

## IIM42652
`spi.go` contains the code for the IIM42652 IMU sensor. This sensor is connected to the SPI bus of the Raspberry Pi.
The `init` function is where the sensor is configured and must be call by the `data logger`.
It is also where all the code to access and modify the sensor registers is located.

### Accelerometer
This where to code to read the accelerometer data is located. call `GetAcceleration` to get the acceleration data.

### Gyroscope
This where to code to read the gyroscope data is located. call `GetGyroscopeData` to get the gyroscope data.

### Temperature
This where to code to read the temperature data is located. call `GetTemperature` to get the temperature data.

`GetAcceleration`, `GetGyroscopeData` and `GetTemperature` are called by the `data logger` every 10ms.