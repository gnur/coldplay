import time
import machine
from machine import Pin

d1 = Pin(5, Pin.OUT)
d2 = Pin(5, Pin.IN)
led = Pin(2, Pin.OUT)

while True:
    print("ok")
    time.sleep(1)
    led(0)
    time.sleep(1)
    led(1)
    time.sleep(1)
