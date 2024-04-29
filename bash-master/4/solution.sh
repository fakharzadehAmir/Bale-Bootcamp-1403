function check_disk_usage() {
    # Check for threshold and if it is not given use default threshold
    threshold=$1
    if [[ threshold -eq "" ]]; then
        threshold=$(( 80 ))
    fi

    # Get information about disk usage
    info=$( echo "$( df -k . )" | awk "NR == 2" ) 


    file_system=$( echo "$info" | awk '{print $1}' )
    available=$( echo "$info" | awk '{print $4}' )
    used=$( echo "$info" | awk '{print $3}' )

    # Evaluate the usage of disk
    disk_usage=$( expr $used \* 100 / $available )

    # Printing the current date & time
    current_datetime=$(date +"%Y-%m-%d %H:%M:%S.%3N")
    # Comparing disk usage and the threshold to warn the admin
    if [[ $disk_usage -gt $threshold ]]; then
        echo "Date and Time: $current_datetime"
        echo WARNING: The partition \""$file_system"\" has used "$disk_usage"% of total available space
    fi
}

check_disk_usage $1


# For monitoring the usage of disk every 10 seconds you can use below command 
# watch -n 10 bash solution.sh <Threshold>