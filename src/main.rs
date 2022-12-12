use std::thread;
use std::time::Duration;

use i2cdev::core::*;
use i2cdev::linux::{LinuxI2CDevice, LinuxI2CError};

const TFLUNA_ADDR: u16 = 0x10;

const MODE_ADDR: u8 = 0x23;
const TRIGGER_MODE: [u8; 1] = [0x01];

const TRIGGER_ADDR: u8 = 0x24;

const SAMPLES: u8 = 100;

fn i2cfun() -> Result<(), LinuxI2CError> {
    let mut dev = LinuxI2CDevice::new("/dev/i2c-1", TFLUNA_ADDR)?;

    let nc = nats::connect("uranus")?;
    nc.publish("coldplay.connection", "connecting")?;

    //set up the TF Luna to only measure when asked for enhanced accuracy
    dev.smbus_write_block_data(MODE_ADDR, &TRIGGER_MODE)?;

    let mut last_measurement = 0.0;
    loop {
        let mut sum: f64 = 0.0;
        let mut reads = 0;
        //do SAMPLES measurements
        for _ in 0..SAMPLES {
            thread::sleep(Duration::from_millis(10));

            dev.smbus_write_block_data(TRIGGER_ADDR, &TRIGGER_MODE)?;

            thread::sleep(Duration::from_millis(10));

            let read = dev.smbus_read_i2c_block_data(0x00, 0x02);
            match read {
                Ok(buf) => {
                    let mut dist: f64 = buf[1] as f64 * 256.0;
                    dist += buf[0] as f64;
                    sum += dist;
                    reads += 1;
                }
                Err(e) => println!("Failure: {}", e),
            }
        }
        //a measurement failed, discard this result
        if reads != SAMPLES {
            continue;
        }

        let avg: f64 = sum / reads as f64;
        let diff = avg - last_measurement;

        if diff.abs() > 0.25 {
            println!("We're moving!");
            //push to nats
        }
        last_measurement = avg;

        println!("avg: {avg}");
    }
    nc.publish("coldplay.venus.connection", "disconnecting")?;
}

fn main() {
    let res = i2cfun();
    match res {
        Ok(_) => println!("Done"),
        Err(err) => println!("Shit: {err}"),
    }
}
