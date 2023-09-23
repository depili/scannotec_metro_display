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

The message length is the length of the data bytes payload and the checksum.


### message types:
* 0x05 - set addresses
  * Payload is 2 bytes with the alternative addresses to listen to
* 0x09 - set RTC
  * Payload is 3 bytes, hours, minutes and seconds to set the time to
* 0x55 U - update display
  * Loads the payload to a 8000 byte buffer waiting to be processed, can be broken up into multiple messages
* 0x56 V - enable display
  * Triggers message processing
  * No timeout on the display
  * No payload, message processing ends at the msg type byte
* 0x57 W - enable display with timeout
  * Triggers message processing
  * Clears the display once a longish timeout expires
  * No payload, message processing ends at the msg type byte
* 0x81 - Ping?
  * Send ACIA flags
  * Triggers message processing
  * No payload, message processing ends at the msg type byte
* 0x87 - RTC related?

Other messages will raise a error flag, but might still write to the 8000 byte buffer, needs to be verified...

### content control characters

* 0x09 scroll text
  * two hex chars following, offset to start the scrolling in hex, - 0x4c, in pixels
* 0x0e small font
  * Confirmed
* 0x0f large font
  * Confirmed
* 0x11 set blink
  * Confirmed
* 0x12 reset blink
  * Confirmed
* 0x13 Enable dynamic content for the row
  * Scroller doesn't scroll or timers change unless this is present in the row data
* 0x14 timed messages
  * Two hex chars following.
  * Need to have multiple sections of different texts
  * Can have different content on each of them
* 0x15 4 char string substition
  * No idea what the contents are supposed to be, maybe temperature reading?
* 0x16 time of day substitution
  * 12:34
* 0x1b <ascii number> set active row
  * Rows start from 0x30 ->


- set_address: return
- V -> process all
- time_set: return
- rtc_something does something: return
- !W -> return



- Jumpperit sarjapiirin läheellä valitsevat RX/TX tyypin