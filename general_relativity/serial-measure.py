######################################################
# Copyright (c) 2021 Maker Portal LLC
# Author: Joshua Hrisko
######################################################
#
# TF-Luna Mini LiDAR wired to a Raspberry Pi via UART
# --- Real-time ranging with signal strength indicator
#
#
######################################################
#
import serial, time

import logging

logging.basicConfig(format="%(asctime)s - %(message)s", level=logging.INFO)
logging.info("starting")


# geel 3 TX
# wit 2 RX
# grijs 5 config / 3v3
#
############################
# Serial Functions
############################
#
def read_tfluna_data():
    while True:
        counter = ser.in_waiting  # count the number of bytes waiting to be read
        bytes_to_read = 9
        if counter > bytes_to_read - 1:
            bytes_serial = ser.read(bytes_to_read)  # read 9 bytes
            ser.reset_input_buffer()  # reset buffer

            if (
                bytes_serial[0] == 0x59 and bytes_serial[1] == 0x59
            ):  # check first two bytes
                distance = (
                    bytes_serial[2] + bytes_serial[3] * 256
                )  # distance in next two bytes
                strength = (
                    bytes_serial[4] + bytes_serial[5] * 256
                )  # signal strength in next two bytes
                temperature = (
                    bytes_serial[6] + bytes_serial[7] * 256
                )  # temp in next two bytes
                temperature = (temperature / 8) - 256  # temp scaling and offset

                return distance, strength, temperature


def set_samp_rate(samp_rate=100):
    ##########################
    # change the sample rate
    samp_rate_packet = [0x5A, 0x06, 0x03, samp_rate, 00, 00]  # sample rate byte array
    ser.write(samp_rate_packet)  # send sample rate instruction
    return


def set_output_format(mm=False, cm=False):
    if mm == cm:
        return

    output_packet = [
        0x5A,
        0x05,
        0x05,
        0x01,  # cm output
        00,
    ]

    if mm:
        # change the output format
        output_packet = [
            0x5A,
            0x05,
            0x05,
            0x06,  # mm output
            00,
        ]  # sample rate byte array
    ser.write(output_packet)  # send sample rate instruction
    return


def get_version():
    ##########################
    # get version info
    info_packet = [0x5A, 0x04, 0x14, 0x00]

    ser.write(info_packet)  # write packet
    time.sleep(0.1)  # wait to read
    bytes_to_read = 30  # prescribed in the product manual
    t0 = time.time()
    while (time.time() - t0) < 5:
        counter = ser.in_waiting
        if counter > bytes_to_read:
            bytes_data = ser.read(bytes_to_read)
            ser.reset_input_buffer()
            if bytes_data[0] == 0x5A:
                version = bytes_data[3:-1].decode("utf-8")
                print("Version -" + version)  # print version details
                return
            else:
                ser.write(info_packet)  # if fails, re-write packet
                time.sleep(0.1)  # wait


def set_baudrate(baud_indx=4):
    ##########################
    # get version info
    baud_hex = [
        [0x80, 0x25, 0x00],  # 9600
        [0x00, 0x4B, 0x00],  # 19200
        [0x00, 0x96, 0x00],  # 38400
        [0x00, 0xE1, 0x00],  # 57600
        [0x00, 0xC2, 0x01],  # 115200
        [0x00, 0x84, 0x03],  # 230400
        [0x00, 0x08, 0x07],  # 460800
        [0x00, 0x10, 0x0E],
    ]  # 921600
    info_packet = [
        0x5A,
        0x08,
        0x06,
        baud_hex[baud_indx][0],
        baud_hex[baud_indx][1],
        baud_hex[baud_indx][2],
        0x00,
        0x00,
    ]  # instruction packet

    prev_ser.write(info_packet)  # change the baud rate
    time.sleep(0.1)  # wait to settle
    prev_ser.close()  # close old serial port
    time.sleep(0.1)  # wait to settle
    ser_new = serial.Serial(
        "/dev/ttyAMA0", baudrates[baud_indx], timeout=1
    )  # new serial device
    if ser_new.isOpen() == False:
        ser_new.open()  # open serial port if not open
    bytes_to_read = 8
    t0 = time.time()
    logging.info("trying to read some bytes")
    while (time.time() - t0) < 15:
        counter = ser_new.in_waiting
        if counter > bytes_to_read:
            bytes_data = ser_new.read(bytes_to_read)
            ser_new.reset_input_buffer()
            if bytes_data[0] == 0x5A:
                indx = [
                    ii
                    for ii in range(0, len(baud_hex))
                    if baud_hex[ii][0] == bytes_data[3]
                    and baud_hex[ii][1] == bytes_data[4]
                    and baud_hex[ii][2] == bytes_data[5]
                ]
                print("Set Baud Rate = {0:1d}".format(baudrates[indx[0]]))
                time.sleep(0.1)
                return ser_new
            else:
                ser_new.write(info_packet)  # try again if wrong data received
                time.sleep(0.1)  # wait 100ms
                continue
    logging.error("Unable to read bytes in 15 seconds")


#
############################
# Configurations
############################
#
baudrates = [9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600]  # baud rates
prev_indx = 4  # previous baud rate index (current TF-Luna baudrate)
prev_ser = serial.Serial(
    "/dev/ttyAMA0", baudrates[prev_indx], timeout=0
)  # mini UART serial device
if prev_ser.isOpen() == False:
    prev_ser.open()  # open serial port if not open
baud_indx = 4  # baud rate to be changed to (new baudrate for TF-Luna)
logging.info("setting baudrate")


ser = set_baudrate(baud_indx)  # set baudrate, get new serial at new baudrate
logging.info("setting sample rate")
set_samp_rate(1)  # set sample rate 1-250
set_output_format(mm=True)
logging.info("getting version")
get_version()  # print version info for TF-Luna

print("Starting Ranging...")
while True:
    distance, strength, temperature = read_tfluna_data()  # read values
    print(distance, temperature)

ser.close()  # close serial port
