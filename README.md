# photoframe
Digital Photoframe USB source.

This project makes use of a Raspberry Pi Zero to act as a USB source for a Digital Photo Frame.
See this article here, https://www.raspberrypi.org/magpi/pi-zero-w-smart-usb-flash-drive

This micro-service will periodically download images from the selected provider and overlay them with the current weather forecast.  The overlayed images will then be placed in the USB source folder and the USB source will be refreshed to force the Digital Photo Frame to refresh the images.

## Quick Start

1. Flash the latest Raspian 32-bit Lite version to a SD card.
1. Insert it into your Raspberry Pi Zero and boot.
1. Follow on-screen instructions to set up and get on the network.
1. Execute ``` sudo nano /boot/config.txt ``` and add ``` dtoverlay=dwc2 ``` near the bottom of the file (before the sections start e.g. [cm4] or [all])
1. Execute ``` sudo reboot now ```
1. Execute ``` sudo dd if=/dev/zero of=/home/pi/usb.img bs=1M count=1000 ``` to create a 1Gb drive to hold the USB files.
1. Execute ``` sudo mkdosfs -F 32 /home/pi/usb.img ``` to format the drive as Fat32
1. Execute ``` sudo modprobe g_multi file=/home/pi/usb.img stall=0 removeable=1 ``` to set the zero as a USB drive.
1. Execute ``` sudo modprobe -r g_multi ``` to reset.

## To mount the drive as a local media directory
1. Execute ``` sudo nano /etc/fstab ```
1. Add the text ``` /home/pi/usb.img      /media/usb      vfat    users,umask=000   0       2 ``` to the bottom of the file and save.
1. Execute ``` sudo mount -a ```
