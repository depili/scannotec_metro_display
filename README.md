# Scannotec metro display


## Message frame

1. 0x82
2. 0x00
3. address byte
4. msg type byte
  * Message might end here
5. length low byte
6. length high byte
7. data bytes
8. checksum
  * Just a rolling 8bit sum of all bytes from step 5


### message types:
* V - enable display
* W - clear display
* 0x05 - set addresses
* 0x09 - set RTC
* 0x81 - something?
* 0x87 - RTC related

any other type will set display contents

### content control characters

* 0x09 goto offset?
* 0x0e small font
* 0x0f large font
* 0x11 set blink
* 0x12 reset blink
* 0x13 goto row start
* 0x14 delay
* 0x15 4 char string substitution
* 0x16 time of day substitution
* 0x1b <byte> set active row