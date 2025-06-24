# SH1107 (i2C) OLED driver in pure Go
Exactly what it does on the tin. Allows you to connect a SH1107 OLED display to your Raspberry Pi (or other computer with an i2C bus) and control it in Golang.

# Why?
Because I had some "fun" with one of these. A 128x128 monochrome OLED display that's based on the SH1107 controller.

![51+LNLog8lL _SL1050_](https://github.com/user-attachments/assets/9b732c68-0b71-4f68-a82c-126e13a7239c)

There weren't any native, easy-to-use Go libraries out there for this particular display. And despite my efforts, not even Python wanted to play with it.

# How?
Thanks to the demo code provided by seeed studio and waveshare (plus a little ChatGPT and Copilot magic) I ported the necessary driver functions needed to
run a SH1107 at a fast enough speed.

# Ok, how do I use it?
Add the sh1107.go as a new library to your project, then import it as necessary. See some demo code here:
