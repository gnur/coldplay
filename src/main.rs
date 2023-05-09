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

const RESET_ADDR: u8 = 0x21;
const RESET_VAL: [u8; 1] = [0x02];

const SAMPLES: u16 = 500;

#[derive(Serialize, Deserialize, Debug)]
struct Measerement {
    height: f64,
    temperature: f64,
    readsWithoutFault: u64,

    #[serde(with = "time::serde::rfc3339")]
    pub timestamp: time::OffsetDateTime,
}

fn i2cfun() -> Result<(), LinuxI2CError> {
    let mut dev = LinuxI2CDevice::new("/dev/i2c-1", TFLUNA_ADDR)?;

    let nc = nats::connect("uranus")?;
    nc.publish("paradise.connection", "connecting")?;

    //set up the TF Luna to only measure when asked for enhanced accuracy
    dev.smbus_write_block_data(MODE_ADDR, &TRIGGER_MODE)?;

    let sleeptime = Duration::from_millis(30);
    let mut lastresult: f64 = 0.0;
    let mut reads_without_fault: u64 = 0;

    loop {
        let mut sum: f64 = 0.0;
        let mut temperature_sum: f64 = 0.0;
        let mut reads = 0;

        //do SAMPLES measurements
        for _ in 0..SAMPLES {
            thread::sleep(sleeptime);

            dev.smbus_write_block_data(TRIGGER_ADDR, &TRIGGER_MODE)?;

            thread::sleep(Duration::from_millis(15));

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

        let temperature_average: f64 = temperature_sum / reads as f64;

        //average of last SAMPLES reads
        let avg: f64 = sum / reads as f64;

        //elevator max observed speed is about +- 5 at this point in the script
        let p = Measerement {
            height: avg,
            readsWithoutFault: reads_without_fault,
            temperature: temperature_average,
            timestamp: time::OffsetDateTime::now_utc(),
        };
        let p = serde_json::to_string(&p).unwrap();
        let res = nc.publish("coldplay.measurement", p);
        match res {
            Ok(()) => {}
            Err(e) => println!("Publishing to NATS failed: {e}"),
        }

        println!("avg={avg}");

        if lastresult == avg {
            //exactly same result, seems suspicous, lets take a break

            nc.publish("paradise.reset", "triggering reset")?;

            thread::sleep(Duration::from_secs_f64(5.0));

            dev.smbus_write_block_data(RESET_ADDR, &RESET_VAL)?;

            thread::sleep(Duration::from_secs_f64(15.0));
            reads_without_fault = 0;
        }
        reads_without_fault += 1;
        lastresult = avg;
    }
}

fn main() {
    let res = i2cfun();
    match res {
        Ok(_) => println!("Done"),
        Err(err) => println!("Shit: {err}"),
    }
}
