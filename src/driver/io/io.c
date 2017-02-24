#include <comedilib.h>
#include <assert.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <netdb.h>
#include <stdio.h>
#include <pthread.h>


#include "io.h"
#include "channels.h"
#include <stdio.h>
#include "con_load.h"



static comedi_t *it_g = NULL;


int io_init(void) {

    it_g = comedi_open("/dev/comedi0");

    if (it_g == NULL) {
        return 0;
    }

    int status = 0;
int i = 0;
    for (i; i < 8; i++) {
        status |= comedi_dio_config(it_g, PORT_1_SUBDEVICE, i + PORT_1_CHANNEL_OFFSET, PORT_1_DIRECTION);
        status |= comedi_dio_config(it_g, PORT_2_SUBDEVICE, i + PORT_2_CHANNEL_OFFSET, PORT_2_DIRECTION);
        status |= comedi_dio_config(it_g, PORT_3_SUBDEVICE, i + PORT_3_CHANNEL_OFFSET, PORT_3_DIRECTION);
        status |= comedi_dio_config(it_g, PORT_4_SUBDEVICE, i + PORT_4_CHANNEL_OFFSET, PORT_4_DIRECTION);
    }

    return (status == 0);
}

static int sockfd;
static pthread_mutex_t sockmtx;

int sim_init(void) {
        char ip[16] = {0};
        char port[8] = {0};
        con_load("simulator.con",
            con_val("com_ip",   ip,   "%s")
            con_val("com_port", port, "%s")
        )
        
        pthread_mutex_init(&sockmtx, NULL);
    
        sockfd = socket(AF_INET, SOCK_STREAM, 0);
        assert(sockfd != -1 && "Unable to set up socket");

        struct addrinfo hints = {
            .ai_family      = AF_UNSPEC, 
            .ai_socktype    = SOCK_STREAM, 
            .ai_protocol    = IPPROTO_TCP,
        };
        struct addrinfo* res;
        getaddrinfo(ip, port, &hints, &res);

        int fail = connect(sockfd, res->ai_addr, res->ai_addrlen);
        assert(fail == 0 && "Unable to connect to simulator server");

        freeaddrinfo(res);

        send(sockfd, (char[4]) {0}, 4, 0);
	
	return 1;
}


void io_set_bit(int channel) {
    comedi_dio_write(it_g, channel >> 8, channel & 0xff, 1);
}


void io_clear_bit(int channel) {
    comedi_dio_write(it_g, channel >> 8, channel & 0xff, 0);
}


void io_write_analog(int channel, int value) {
    comedi_data_write(it_g, channel >> 8, channel & 0xff, 0, AREF_GROUND, value);
}


int io_read_bit(int channel) {
    unsigned int data = 0;
    comedi_dio_read(it_g, channel >> 8, channel & 0xff, &data);
    return (int)data;
}


int io_read_analog(int channel) {
    lsampl_t data = 0;
    comedi_data_read(it_g, channel >> 8, channel & 0xff, 0, AREF_GROUND, &data);
    return (int)data;
}



