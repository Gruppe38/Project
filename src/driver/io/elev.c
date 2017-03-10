#include "elev.h"

#include <assert.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <netdb.h>
#include <stdio.h>
#include <pthread.h>

#include "channels.h"
#include "io.h"
#include "con_load.h"

#define MOTOR_SPEED 2800


static const int lamp_channel_matrix[N_FLOORS][N_BUTTONS] = {
    {LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
    {LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
    {LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
    {LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
};


static const int button_channel_matrix[N_FLOORS][N_BUTTONS] = {
    {BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
    {BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
};



static elev_type elevatorType = ET_Simulation;
static int sockfd;
static pthread_mutex_t sockmtx;

void elev_init() {
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
}




void elev_set_motor_direction(int dirn) {
    switch(elevatorType) {
    case ET_Comedi:
        break;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {1, dirn}, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        break;
    }
}


void elev_set_button_lamp(int button, int floor, int value) {
    switch(elevatorType) {
    case ET_Comedi:
        break;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {2, button, floor, value}, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        break;
    }
}


void elev_set_floor_indicator(int floor) {
    switch(elevatorType) {
    case ET_Comedi:
        break;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {3, floor}, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        break;
    }
}


void elev_set_door_open_lamp(int value) {
    switch(elevatorType) {
    case ET_Comedi:
        break;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {4, value}, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        break;
    }
}


int elev_get_button_signal(int button, int floor) {
    switch(elevatorType) {
    case ET_Comedi:
        return 0;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {6, button, floor}, 4, 0);
        char buf[4];
        recv(sockfd, buf, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        return buf[1];
    }
    return 0;
}


int elev_get_floor_sensor_signal(void) {
    switch(elevatorType) {
    case ET_Comedi:
        return 0;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {7}, 4, 0);
        char buf[4];
        recv(sockfd, buf, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        return buf[1] ? buf[2] : -1;
    }
    return 0;
}




int elev_get_obstruction_signal(void) {
    switch(elevatorType) {
    case ET_Comedi:
        return 0;
    case ET_Simulation:
        pthread_mutex_lock(&sockmtx);
        send(sockfd, (char[4]) {9}, 4, 0);
        char buf[4];
        recv(sockfd, buf, 4, 0);
        pthread_mutex_unlock(&sockmtx);
        return buf[1];
    }
    return 0;
}
