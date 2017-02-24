#pragma once

typedef enum {
    ET_Comedi,
    ET_Simulation
} elev_type;

// Returns 0 on init failure
int io_init(void);
int sim_init(void);

void io_set_bit(int channel);
void io_clear_bit(int channel);

int io_read_bit(int channel);

int io_read_analog(int channel);
void io_write_analog(int channel, int value);

