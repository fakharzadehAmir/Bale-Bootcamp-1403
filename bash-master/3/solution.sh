# I don't use first line as the size of given array
declare -A hashmap

array=$( cat $1 | awk 'NR == 2')

# count numbers of occurances
for num in ${array[@]}
do 
    if [[ -n ${hashmap[$num]} ]]; then 
        hashmap[$num]=$((hashmap[$num] + 1))
    else
        hashmap[$num]=1
    fi
done

# find the number with 1 occurances
for num in "${!hashmap[@]}"
do
    if [[ ${hashmap[$num]} -eq 1 ]]; then
        echo $num
        break
    fi
done