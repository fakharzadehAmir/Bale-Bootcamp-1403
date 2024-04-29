# Read the given file 
file=$( cat $1 )


for line in "$file" 
do
    echo "$line" | awk '{ if ( $2 == "error" ) print;}'
done    