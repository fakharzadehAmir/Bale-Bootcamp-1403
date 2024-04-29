# Read the given file 
numbers=$( cat $1 )
# Mask first 3 parts of each line using sed command
for (( i=1; i<=4; i++ ))
do
    number=$( echo "$numbers" | awk "NR == $i" )
    masked_variable=$(echo "$number" | sed 's/[0-9]\+/\****/1; s/[0-9]\+/\****/1; s/[0-9]\+/\****/1')
    echo "$masked_variable"
done

