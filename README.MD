#Momento 3 To Day One Importer

It's a simple tool that helps you migrate from Momento 3 to Day One. It can help you import data **(moments and photos)** from a Momento 3 backup into Day One.

Note that it **WON'T** import any feed moment (except for RSS feed), tag, event, people, etc, into Day One. 

If a moment has more than one photo, only the first photo **MAY**(dont know why) still be with the moment, the others will be imported as sperated entries because Day One CLI does not support multiple photos for one entry.

For all photos added before Momento 3, they will be imported according to their **MODIFIED DATE**, because there is no way to find out which moment they are in. 

Defaultly, it stores moments in your Default Journal.dayone, located here:

`~/Library/Group Containers/5U8NS4GX82.dayoneapp2/Data/Auto Import/Default Journal.dayone`

##Requirements:

- Golang
- Day One for Mac
- Day One CLI
- Momento 3, of course

##Steps:

1. Make a Momento 3 backup and export it to your mac.
2. Unzip that backup, you will get a folder "Backups", then unzip the latest backup data in folder "Data" inside and rename the new folder into "backup".
3. Build this project, or, `go get github.com/int123c/Momento3toDayOne`, and run it in terminal trailing by the path of folder "Backups" like: <br> `> ./Momento3toDayOne -p="/Users/username/Downloads/Backups"` <br> Or, you may put the binary file into "Backups" and run it without augument.
4. Wait for it to finish.

##Attention:

Please **turn of sync** before you do it. (incase you don't like the result)

If you don't want it to store moments into your default journal, edit `main.go` and change `jounalPath` to another `.dayone` file.


**Not guaranteed to work. (at least worked for me)**

