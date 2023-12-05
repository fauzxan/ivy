# ivy 🚀
Implementation of the Integrated shared Virtual memory at Yale. 

# How to run the code

## Setting up the terminals.

Each server, backup, and client will run on a spearate terminal, or even separate systems on the same network. The reboots will work, provided the local DHCP doesn't allocate a new address for some arbitrary reason. This typically shouldn't happen if you didn't close your laptop between the crash and failure, as DHCPs only reassign IPs after a some preset timeout. 

The first server you spin up will be the central manager. 
```shell
go build && ./ivy -cm
```

The second server you spin up will be the backup, provided you pass in the IP of the first server to it. How do you find out the IP of the first server? Easy, just look at the logs on the first terminal:

```shell
There is no backup!
2023/12/05 21:02:34 Central manager!
My IP is:172.23.129.167:35677
Central manager is runnning at IP address:172.23.129.167:35677 <-- Look for this line, copy this onto your clipboard
```
Once you have copied the IP address of the central manager, you may now procees to create:
1. Backup central manager using `go build && ./ivy -cm <IP ADDRESS THAT YOU JUST COPIED>`
2. As many clients as you want on separate terminals using `go build && ./ivy -u <IP ADDRESS THAT YOU JUST COPIED>`


Great! Once you have this setup, we can go over the commands to do the following:

1. Start the never ending cycle of randomly issuing read and write requests. You may stop this behaviour by adjusting the code to do it only once in `main.go`
2. Kill any server. Central or Backup. This can easily be done by pressing ctrl+C in the central manager/ backup manager's terminal.
3. View some important log information like Record cache at the server, or the Page Cache at the clients.

## How to reboot

The code is originally designed to assign a random, unused port number for the first server. This is to save you from the agony of having to choose a port number 🤯🤯🤯

So in order to reboot at the same port number, simply issue the following command:
```shell
go build && ./ivy -r <IP ADDRESS OF THE SERVER THAT JUST WENT DOWN>
```

### What happens when you reboot?

When the central manager (not backup) reboots, it will receive a PING message from the backup, and it will tell all the clients to contact it from here onwards. 

When the backup manager reboots, it will simply wait for the central manager to die in order to do anything meaningful 😆 However, once the central manager dies (process exits), the backup will keep nagging (sending ping message) the central manager to see if it is back alive. Once it is back alive, the backup manager will send over all the details that it had updated when the central was dead, to the central manager.

## Useful commands for the clients (AKA 1,2,3,4...)

"1" will print the clientlist

"2" will print the IP of the server you are currently contacting. This may be the central server, or the backup server. 

"3" will start the never ending cycle of randomly issuing read or write requests. All read or write requests are made on randomized pages 1,2, or 3. You may increase the number of pages by changing the following line in client.ReadRequest() and client.WriteRequest():

```go
 pageid := rand.Intn(3)
```
A read or a write request will be randomly issued once every 15 seconds. This is to prevent the logs from going 💣💣💣

Only way you stop this never ending cycle is by killing the client. 

## Useful commands for the server (AKA 1,2,...)

"1" Will print the clientlist

"2" will print the records that you have in your posession. This is the list of pages, and their respective owners, and the list of IPs that have a copy of it. 

# Some interesting design details

## How do reads happen?
![image](https://github.com/fauzxan/ivy/assets/92146562/38ec64ab-4a91-4d3e-86c1-a452ddb6bda1)


## How do writes happen?
![image](https://github.com/fauzxan/ivy/assets/92146562/65629a1a-7882-428d-9d7b-aa039bd7d8b4)

Writes maintain a queue at the server. As of now, there is no inherent benefit to having a separate WRITE access mode at the client. This will be useful in scenarios where the client performs multiple computations on the page before falling back to write mode. However, the code only handles a single variable update for any page. 

