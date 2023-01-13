use serde::{Deserialize, Serialize};
use std::thread;
use std::time::Duration;
use time;

use i2cdev::core::*;
use i2cdev::linux::{LinuxI2CDevice, LinuxI2CError};

const TFLUNA_ADDR: u16 = 0x10;

const MODE_ADDR: u8 = 0x23;
const TRIGGER_MODE: [u8; 1] = [0x01];

const TRIGGER_ADDR: u8 = 0x24;

const GROUND_FLOOR_HEIGHT: f64 = 2.5;
const FIRST_FLOOR_HEIGHT: f64 = 276.6;
const TOP_FLOOR_HEIGHT: f64 = 544.0;

const SAMPLES: u8 = 50;

#[derive(Serialize, Deserialize, Debug)]
struct Measerement {
    height: f64,
    temperature: f64,

    #[serde(with = "time::serde::rfc3339")]
    pub timestamp: time::OffsetDateTime,
}

fn i2cfun() -> Result<(), LinuxI2CError> {
    let mut dev = LinuxI2CDevice::new("/dev/i2c-1", TFLUNA_ADDR)?;

    let nc = nats::connect("uranus")?;
    nc.publish("coldplay.connection", "connecting")?;

    //set up the TF Luna to only measure when asked for enhanced accuracy
    dev.smbus_write_block_data(MODE_ADDR, &TRIGGER_MODE)?;

    let mut last_measurement = -100.0;
    let mut sleeptime = Duration::from_millis(20);
    let mut moves = 0;

    loop {
        let mut sum: f64 = 0.0;
        let mut temperature_sum: f64 = 0.0;
        let mut reads = 0;
        //do SAMPLES measurements
        for _ in 0..SAMPLES {
            thread::sleep(sleeptime);

            dev.smbus_write_block_data(TRIGGER_ADDR, &TRIGGER_MODE)?;

            thread::sleep(sleeptime);

            let read = dev.smbus_read_i2c_block_data(0x00, 0x06);
            match read {
                Ok(buf) => {
                    let mut dist: f64 = buf[1] as f64 * 256.0;
                    dist += buf[0] as f64;
                    sum += dist;

                    let mut temp: f64 = buf[5] as f64 * 256.0;
                    temp += buf[4] as f64;
                    temp = temp / 100.0;
                    temperature_sum += temp;

                    reads += 1;
                }
                Err(e) => println!("Failure: {}", e),
            }
        }
        //a measurement failed, discard this result
        if reads != SAMPLES {
            continue;
        }

        let temperature_average: f64 = (temperature_sum / reads as f64);

        //average of last SAMPLES reads
        let mut avg: f64 = 545.0 - (sum / reads as f64);

        if last_measurement == -100.0 {
            last_measurement = avg;
        }

        avg = (avg * 2.0 + last_measurement) / 3.0;
        //flip measurements upside down to let them make a bit more sense and take a moving average

        let diff = avg - last_measurement;

        //elevator max observed speed is about +- 5 at this point in the script
        if diff.abs() > 15.0 {
            println!("Skipping because init: {diff}");
            last_measurement = avg;
            //this should only happen on boot, skip processing until measurements are stable
            continue;
        }

        //check if within +- 1 cm of ground floor: 2.5
        //check if within +- 1 cm of middle floor: 276.6
        //check if within +- 1 cm of top floor: 544

        if diff.abs() > 0.25 {
            moves = 5;
            println!("We're moving!");
            //push to nats
            sleeptime = Duration::from_millis(10);

            let p = Measerement {
                height: avg,
                temperature: temperature_average,
                timestamp: time::OffsetDateTime::now_utc(),
            };
            let p = serde_json::to_string(&p).unwrap();
            let res = nc.publish("coldplay.measurement", p);
            match res {
                Ok(()) => {}
                Err(e) => println!("Publishing to NATS failed: {e}"),
            }
        } else {
            sleeptime = Duration::from_millis(20);
            moves -= 1;

            if moves == 0 {
                println!("We've stopped moving");
            }
            if moves == -60 {
                println!("Publishing static state");
                moves = 0;
                //also trigger
                let p = Measerement {
                    height: avg,
                    temperature: temperature_average,
                    timestamp: time::OffsetDateTime::now_utc(),
                };
                let p = serde_json::to_string(&p).unwrap();
                let res = nc.publish("coldplay.measurement", p);
                match res {
                    Ok(()) => {}
                    Err(e) => println!("Publishing to NATS failed: {e}"),
                }
            }
        }

        last_measurement = avg;

        println!("avg: {avg}");
    }
}

fn main() {
    let res = i2cfun();
    match res {
        Ok(_) => println!("Done"),
        Err(err) => println!("Shit: {err}"),
    }
}
