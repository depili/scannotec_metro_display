// Manually cleaned decompiled functions, might have errors in pointer increments, due to trying to clean them up...


// This function does the initial receive handling.
// Baudrate is set to 4800 in our specimen.
// The payload is processed in ACIA_msg_process

void ACIA_ISR_fast(void) {
  byte bVar1;
  enum_acia_flags eVar2;
  enum_acia_msg_type *store_incremented;
  enum_acia_msg_type rx_byte;
  enum_acia_msg_type temp;

  // Check interrupt flag in ACIA ctrl register
  if (IO_ACIA_CTRL & 0x80 == 0) {
    /* Error state, call reset vector */
    reset();
  }
  rx_byte = IO_ACIA_DATA;
  if (ACIA_flags & DO_RX == 0) {
    return;
  }

  /* Check receiving state */
  switch (ACIA_rx_state) {
  case 0:
    // Start byte 1 search
    if (rx_byte != 0x82) {
      ACIA_RX_timeout = 0;
      return;
    }
    break;
  case 1:
    // Start byte 2 search
    if (rx_byte != 0) {
      ACIA_rx_state = 0;
      ACIA_RX_timeout = 0;
      return;
    }
    break;
  case 2:
    // Message address
    addr = rx_byte;
    if (addr != address[0] && addr != address[1] && addr != address[2]) {
      ACIA_rx_state = 0;
      ACIA_RX_timeout = 0;
      return;
    }
    ACIA_incoming_address = addr;
    break;
  case 3:
    // Message type
    switch (rx_byte) {
    case 'V':
      ACIA_FLAGS = 0;
      ACIA_rx_state = 0;
      ACIA_msg_type = 0x03;
      ACIA_RX_timeout = 0;
      return;
    case 'W':
      ACIA_flags = 0;
      ACIA_rx_state = 0;
      ACIA_msg_type = 0x04;
      ACIA_RX_timeout = 0;
      return;
    case 0x81:
      if (rx_byte == address[0]) {
        IO_ACIA_DATA = ACIA_flags;
        ACIA_flags |= PINGED;
        ACIA_some_counter = 0;
      }
      ACIA_rx_state = 0;
      ACIA_RX_timeout = 0;
      return;
    case 0x87:
      ACIA_reset_ptrs();
      ACIA_msg_type = 0x87;
      ACIA_checksum = 0;
      ACIA_flags = 0
      ACIA_rx_state = 0;
      ACIA_RX_timeout = 0;
      return;
    default:
      ACIA_reset_ptrs();
      ACIA_msg_type = rx_byte;
      ACIA_checksum = 0;
      ACIA_rx_state += 1;
      ACIA_RX_timeout = 0x32;
      return;
    }
  case 4:
    // Length high byte
    if (rx_byte > 7) {
      eVar2 = ERR_LENGTH;
      err_frame(eVar2);
      return;
    }
    ACIA_msg_len = rx_byte << 8;
    break;
  case 5:
    // Length low byte
    ACIA_msg_len += rx_byte;
    ACIA_rx_state = 0x7f;
    break;
  case 0x80:
    // Payload receive state
    if (ACIA_msg_len == 1) {
      if (rx_byte == ACIA_checksum) {
        // Checksum verification
        if (ACIA_msg_type == U_update_display?) {
          ACIA_some_counter = 0;
          ACIA_flags = DO_RX;
          ACIA_rx_state = 0;
          ACIA_RX_timeout = 0;
          return;
        }
        ACIA_flags = 0
        ACIA_rx_state = 0;
        ACIA_RX_timeout = 0;
        return;
      }
      eVar2 = ERR_CHECKSUM;
      err_frame(eVar2);
      return;
    } else {
      *ACIA_data_store_ptr = rx_byte;
      ACIA_data_store_ptr++;
      ACIA_msg_len--;
      if (ACIA_data_store_ptr >= &ACIA_bufer_end) {
        eVar2 = ERR_OVERRUN;
        err_frame(eVar2);
        return;
      }
      ACIA_checksum = rx_byte + ACIA_checksum;
      ACIA_RX_timeout = 0x32;
      return;
    }
  default:
    reset();
  }
  ACIA_rx_state++
  ACIA_RX_timeout = 0;
  if (ACIA_rx_state != 0) {
    ACIA_RX_timeout = 0x32;
  }
  return;
}

void err_frame(byte eVar2) {
  ACIA_flags = eVar2 | (DO_RX|ERR_FRAME);
  ACIA_rx_state = 0;
  ACIA_RX_timeout = 0;
  return;
}




// Data is received by ISR ACIA_ISR_fast
//
// This function processes the message type and payload

void ACIA_msg_process(void)

{
  byte delay2;
  byte delay1;

  // Check ACIA_FLAGS
  //
  // Only if all flags other than 0x02 are unset do we process the message
  if ((ACIA_flags & 0x02) != 0) {
    return;
  }

  switch (ACIA_msg_type) {
  case set_addresses:
    address._1_2_ = *(undefined2 *)ACIA_data_read_ptr1?;
    ACIA_set_flags(DO_RX);
    return;
  case time_set:
    // Message type 0x09 - set RTC time
    rtc_time._0_2_ = *(undefined2 *)ACIA_data_read_ptr1?;
    rtc_time.seconds = ACIA_data_read_ptr1?[2];
    rtc_register = 0x80;
    msg_type_9_related4 = 0;
    ACIA_set_flags(DO_RX);
    return;
  case rtc_something:
    IO_RTC_ctrl = 0x3f;
    delay1 = 0xc;
    do {
      delay2 = 0;
      do {
        delay2 -= 1;
      } while (delay2 != 0);
      delay1 -= 1;
    } while (delay1 != 0);
    IO_RTC_ctrl = 0x37;
    ACIA_set_flags(DO_RX);
    return;
  case V_enable_display:
    mem_clear_flag = false;
    break;
  case W_clear_display:
    mem_clear_flag = true;
    ACIA_some_counter = 0;
    break;
  default:
    ACIA_flags = DO_RX|ERR_MSG_TYPE|ERR_FRAME;
    return;
  }
  msg_timer = 0;
  output_flip_inhibit = 0xff;
  PIO_2_A_output_state? |= 0x80;
  IO_PIA_2_data_A_address = PIO_2_A_output_state?;
  ACIA_set_row_contents();
  ACIA_reset_ptrs();
  mem_zero_0x100((short *)&rows_bitmaps);
  active_row = 0;
  row_special_chars((char *)row_0_str,rows_bitmaps.row0_a.bitmap);
  mem_cpy_0x100((byte *)&rows_bitmaps,rows_bitmaps.row0_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row1_a);
  active_row = 1;
  row_special_chars((char *)row_1_str,rows_bitmaps.row1_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row1_a.start,rows_bitmaps.row1_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_2_a);
  active_row = 2;
  row_special_chars((char *)row_2_str,rows_bitmaps.row_2_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_2_a.start,rows_bitmaps.row_2_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_3_a);
  active_row = 3;
  row_special_chars((char *)row_3_str,rows_bitmaps.row_3_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_3_a.start,rows_bitmaps.row_3_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_4_a);
  active_row = 4;
  row_special_chars((char *)row_4_str,rows_bitmaps.row_4_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_4_a.start,rows_bitmaps.row_4_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_5_a);
  active_row = 5;
  row_special_chars((char *)row_5_str,rows_bitmaps.row_5_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_5_a.start,rows_bitmaps.row_5_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_6_a);
  active_row = 6;
  row_special_chars((char *)row_6_str,rows_bitmaps.row_6_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_6_a.start,rows_bitmaps.row_6_b.start);
  mem_zero_0x100((short *)&rows_bitmaps.row_7_a);
  active_row = 7;
  row_special_chars((char *)row_7_str,rows_bitmaps.row_7_a.bitmap);
  mem_cpy_0x100(rows_bitmaps.row_7_a.start,rows_bitmaps.row_7_b.start);
  output_flip_inhibit = 0;
  ACIA_set_flags(DO_RX);
  return;
}




void ACIA_set_row_contents(void)

{
  byte bVar1;
  byte bVar2;
  byte *src;
  byte *str_ptr;
  byte *dst;
  byte *pbVar3;
  short len_remaining;
  char chr;
  byte *read_ptr;
  byte cVar1;

  rows_data_clear();
  active_row = 0;
  set_row_contents_state? = 0;
  read_ptr = ACIA_data_read_ptr1?;
  len_remaining = ACIA_content_len;
  for (len = ACIA_content_len; len > 0; len--) {
    chr = *read_ptr
    read_ptr++;
    if (chr < 0x20) {
      switch (chr) {
      case 0x1b:
        set_row_contents_state? = 0;
        store_active_row_ptr(dst);
        read_ptr++;
        active_row = *read_ptr - 0x30;
        switch (active_row) {
        case 0:
          dst = row_0_str;
        case 1:
          dst = row_1_str;
        case 2:
          dst = row_2_str;
        case 3:
          dst = row_3_str;
        case 4:
          dst = row_4_str;
        case 5:
          dst = row_5_str;
        case 6:
          dst = row_6_str;
        case 7:
          dst = row_7_str;
        }
        len -= 1;
        read_ptr++;
        continue;
      case 0x13:
        set_row_contents_state? = 0x13;
        break;
      case 0x09:
      case 0x14:
        pbVar3 = dst + 1;
        *dst = chr;
        bVar2 = (chr < 0x0) << 3;
        bVar1 = hex_to_int((char *)read_ptr);
        if ((bVar2 >> 3 & 1) == 0) {
          row_contents_related = bVar1 << 4;
          bVar2 = ((char)row_contents_related < '\0') << 3;
          bVar1 = hex_to_int((char *)read_ptr);
          if (!(bool)(bVar2 >> 3 & 1)) {
            chr = bVar1 | row_contents_related;
            len -= 2;
            dst++;
            break;
          }
        }
        else {
          read_ptr++;
        }
        len -= 2;
        dst = pbVar3 + -1;
        continue;
      case 0x15:
        *dst = 0x15;
        dst++;
        if (set_row_contents_state? != 0) {
          chr = ' ';
          dst[0] = ' ';
          dst[1] = ' ';
          dst[2] = ' ';
          dst = dst + 3;
          break;
        }
        continue;
      case 0x16:
        *dst = 0x16;
        dst++;
        if (set_row_contents_state? == 0){
          continue;
        }
        chr = ' ';
        dst[0] = ' ';
        dst[1] = ' ';
        dst[2] = ' ';
        dst[3] = ' ';
        dst = dst + 4;
        break;
      case 0x0e:
      case 0x0f:
      case 0x11:
      case 0x12:
        break;
      default:
        read_ptr++;
        continue;
      }

    }
    *dst = chr;
    read_ptr++;
    dst++;
  }
  rows_str_end_ptr[active_row] = dst;
  return;
}
