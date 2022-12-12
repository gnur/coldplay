import TFmini_I2C

a = TFmini_I2C.TFminiI2C(1, 0x10)

print(a.readAll())
