#!/bin/bash

SERVICE_IP=$(kubectl get svc nginx-service -o jsonpath='{.spec.clusterIP}')
END_TIME=$((SECONDS+60))
OUTPUT_FILE="test.txt"

# Clear the output file before starting the test
> $OUTPUT_FILE

while [ $SECONDS -lt $END_TIME ]; do
    RESPONSE=$(curl -s http://$SERVICE_IP)
    if [ $? -ne 0 ]; then
        echo "$(date) - Request failed" >> $OUTPUT_FILE
        exit 1
    else
        # echo "$(date) - Request succeeded: $RESPONSE" >> $OUTPUT_FILE
        echo "$(date) - Request succeeded" >> $OUTPUT_FILE
    fi
    sleep 1
done

echo "$(date) - Session remained uninterrupted" >> $OUTPUT_FILE