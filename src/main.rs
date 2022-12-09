use rppal::gpio::{Gpio, InputPin, OutputPin};
use std::thread;
use std::time::{Duration, Instant};

const GPIO_TRIG: u8 = 14;
const GPIO_ECHO: u8 = 15;

const SONIC_SPEED: f64 = 0.034; // cm/micro second
pub struct Sonar {
    trig: OutputPin,
    echo: InputPin,
}

impl Sonar {
    pub fn new() -> Option<Sonar> {
        let gpio = match Gpio::new() {
            Ok(gpio) => gpio,
            Err(e) => {
                println!("{:?}", e);
                return None;
            }
        };
        let trig = match gpio.get(GPIO_TRIG) {
            Ok(pin) => {
                let output = pin.into_output();
                output
            }
            Err(e) => {
                println!("{:?}", e);
                return None;
            }
        };
        let echo = match gpio.get(GPIO_ECHO) {
            Ok(pin) => {
                let input = pin.into_input();
                input
            }
            Err(e) => {
                println!("{:?}", e);
                return None;
            }
        };
        Some(Sonar { echo, trig })
    }
    /// Returns a distance sample.
    /// Will return -1 if something was a foot
    pub fn get_distance(&mut self) -> f64 {
        self.trig.set_low();
        thread::sleep(Duration::from_micros(2));
        self.trig.set_high();
        thread::sleep(Duration::from_micros(10));
        self.trig.set_low();
        let mut init = Instant::now();
        let mut start = Instant::now();
        let mut duration = Duration::new(0, 0);
        while self.echo.is_low() {
            start = Instant::now();
            if init.elapsed().as_millis() > 30 {
                println!("is_low: {}", init.elapsed().as_millis());
                return -1.0;
            }
        }
        init = Instant::now();
        while self.echo.is_high() {
            duration = start.elapsed();
            if init.elapsed().as_millis() > 30 {
                println!("is_high: {}", init.elapsed().as_millis());
                return -1.0;
            }
        }
        let micros = duration.as_micros();
        let distance = (SONIC_SPEED * micros as f64) / 2.0;
        return distance;
    }
}

fn main() {
    let mut so = Sonar::new().unwrap();
    let dist = so.get_distance();

    println!("{:.2} cm", dist);
}
