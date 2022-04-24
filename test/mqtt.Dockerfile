FROM eclipse-mosquitto:2
ADD mosquitto.conf /mosquitto/config/
ADD mqpass /mosquitto/config/