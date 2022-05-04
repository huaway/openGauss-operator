echo threshold large:  > cost.txt
cat threshold-sla-large.txt |
grep '95th percentile' |
awk '{print $3}' |
awk '{a+=$1;b+=1}END{print a/b}' >> cost.txt

echo threshold small: >> cost.txt
cat threshold-sla-small.txt |
grep '95th percentile' |
awk '{print $3}' |
awk '{a+=$1;b+=1}END{print a/b}' >> cost.txt

echo load prediction: >> cost.txt
cat loadprediction-sla.txt |
grep '95th percentile' |
awk '{print $3}' |
awk '{a+=$1;b+=1}END{print a/b}' >> cost.txt