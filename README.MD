# B2 Watcher

B2 Watcher is a folder sync utility that will automatically upload files to the Backblaze B2 service. [Backblaze B2]

# Why?
If you are like me then you have some servers running at your home or in a private cloud somewhere. With all that work you put in you better have backups! But backing up offsite can get pricey quick. Enter Backblaze B2, the cheapest ($.005/GB) storage out there. I can fully backup my VMs and data (~500GB) offsite now for a few dollars a month. 

# History
I originally wrote a backup app in python using sqlite but this started creating problems. First was I was backing up data from a different server which was causing a lot of network congestion due to the SHA1 calcs and actual transfers. I couldn't get python 3 to run correctly on my NAS so I couldn't run it directly. Since I started writing go a few months ago I thought this would be a good first project.

# What works

  - Use a yaml config file to define folders to scan/monitor, times, b2 info etc..
  - Scanning a folder(s) and comparing it to what's already in B2 and copying only what's missing
  - Set a time in config file to continually monitor folder for new files and upload automatically (daemon mode).
  - Specify multiple folders to upload/monitor in the same config file

# What Doesn't work/Todo
  - More error corrections (ie. failed file upload, failed large files, etc..)
  - API to get status and pass/fail information
  - Updates via email or ... upon sucess/fail etc...

# Cool! How do I use it?
Well I will have some pre-compiled binaries shortly but for now you can simply import it into your go environemnt and build it yourself. 
 > I'll set up an internal Jenkins build to do the compile stuff soon! Promise!

From there you just need to supply your config file and you are off and running.. err... backuping..

```sh
$ b2watcher -config=myconfigfile.yml
```
### Yaml Config File Sample

```yaml
account_id: ********
app_key: **************************
con_connections: 4
log_dir: /var/log
folders:
  - bucket_id: **************
    b2_folder: Testing
    root_folder: /backup/folder
    monitor: true
    hour: 0
    minute: 1
    delete_after_upload: false
```

account_id : Your account ID from Backblaze
app_key: Your app key from Backblaze
con_connections: This is the concurrent connections you want. Depending on how much memory you have you can adjust this. Keep in mind that if you are uploading lots of large files you could be using 200Mb per concurrent connection. Keep this in mind when setting this! (Default is 4 if nothing is specified)
log_dir: Where you want to store the log file b2watcher.log.
folders (Array):
&nbsp;&nbsp;&nbsp;&nbsp;bucket_id: Bucket ID you want to store files in from Backblaze
&nbsp;&nbsp;&nbsp;&nbsp;b2_folder: Top level folder to store files in on Backblaze
&nbsp;&nbsp;&nbsp;&nbsp;root_folder: Folder you want to sync/monitor (leave off trailing \)
&nbsp;&nbsp;&nbsp;&nbsp;monitor: Continuously monitor the folder for changes (default is false)
&nbsp;&nbsp;&nbsp;&nbsp;hour: Hour(s) to wait between scans for new files
&nbsp;&nbsp;&nbsp;&nbsp;minute: Minute(s) to wait between scans
&nbsp;&nbsp;&nbsp;&nbsp;delete_after_upload: Delete the file after upload has been verified (Currently not programmed!)

# Contact

Hopefully you find this little utility useful. If you have any issues please submit a bug or you can email me [matt@toledomaker.com][email]

# Enjoy!
 
[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)


   [Backblaze B2]: <https://www.backblaze.com/b2/cloud-storage.html>
   [email]: <matt@toledomaker.com>

   [PlDb]: <https://github.com/joemccann/dillinger/tree/master/plugins/dropbox/README.md>
   [PlGh]: <https://github.com/joemccann/dillinger/tree/master/plugins/github/README.md>
   [PlGd]: <https://github.com/joemccann/dillinger/tree/master/plugins/googledrive/README.md>
   [PlOd]: <https://github.com/joemccann/dillinger/tree/master/plugins/onedrive/README.md>
   [PlMe]: <https://github.com/joemccann/dillinger/tree/master/plugins/medium/README.md>
   [PlGa]: <https://github.com/RahulHP/dillinger/blob/master/plugins/googleanalytics/README.md>
 

