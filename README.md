# photoframe
Digital Photoframe USB source.

This project makes use of a Raspberry Pi Zero to act as a USB source for a Digital Photo Frame.
See this article here, https://www.raspberrypi.org/magpi/pi-zero-w-smart-usb-flash-drive

This micro-service will periodically download images from the selected provider and overlay them with the current weather forecast.  The overlayed images will then be placed in the USB source folder and the USB source will be refreshed to force the Digital Photo Frame to refresh the images.
