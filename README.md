#nagtool

nagios quick mute, unmute, clean mute by hostname or service.
more func are sting appending

##Function able to use now

```bash
# append -exec means not only list, but also do action
./nagtool -cleanmute #list all the host and service that have been muted but it is OK status
# for example ./nagtool -cleanmute -exec #do unmute the list above
./nagtool -mutehost "hostname\d\d\.com" #list all the host that name is hostname\d\d\.com , able to regex
./nagtool -unmutehost "hostname\d\d\.com" #same above but unmute, it only list the host is not having problem now.
./nagtool -muteservice "SERVICENAME" #list all the service that name is SERVICENAME, able to regex
./nagtool -unmuteservice "SERVICENAME" #same above but unmute, it only list the service is not having problem now.
./nagtool -all #instead of do any sorting, list all the info by json format, not able to do exec
```
- able to use mutehost and muteservice at same time, unmutehost and unmuteservice at same time, but not mute with unmute
- more Function like ack are still implementing
