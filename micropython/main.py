import time
import machine
from machine import Pin

d1 = Pin(5, Pin.OUT)
d2 = Pin(4, Pin.IN)
led = Pin(2, Pin.OUT)
timeout = 30000  # 30.000us * speed of sound = about 10 meters


def measure():
    d1(0)
    time.sleep_us(5)
    d1(1)
    time.sleep_us(10)
    d1(0)
    pulse_time = -1
    try:
        pulse_time = machine.time_pulse_us(d2, 1, timeout)
    except OSError as ex:
        print(ex)

    distance = (pulse_time / 2) / 29.388
    return distance


while True:
    # get 5 measurements and average them out
    total = 0
    valid = 0

    for i in range(5):
        dis = measure()
        if dis == -1:
            continue
        total += dis
        valid += 1

    if valid < 4:
        # measurements failed, ignore these results
        continue

    average = total / valid

    print("Distance: {:2f}".format(average))

    time.sleep(0.1)
