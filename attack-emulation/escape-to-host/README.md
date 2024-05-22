
## Defining Abilities

Since dns-manipulation doesn't pre-exist in caldera abilities so we need to define the abilities by ourselves.

### Create abilities

Commands


```bash
kubectl create -f https://raw.githubusercontent.com/5GSEC/nimbus/main/attack-emulation/escape-to-host/pod.yaml
```

```bash
kubectl get pods nginx 
```

```bash
kubectl exec nginx -- bash -c "echo 'hello world!' >> /test-pd/hello.txt"
```

```bash
kubectl delete pod nginx
```

### Create test pod

![alt text](images/create-test-pod.png)

### Get the pod

![alt text](images/get-pod.png)

### Make changes in hostpath

![alt text](images/make-changes.png)

### Delete test pod

![alt text](images/delete-test-pod.png)


## Create Adversary

- `+` New Profile
- `+` Add Ability

![alt text](images/create-adversary.png)

## Create Operation

- `+` New Operation
- set Adversary

![alt text](images/operation.png)


## Attack Emulation

After creating the operation click on start to start the attack, optionally you can also check locally in your terminal that whether the caldera agent is working as expected or not.

![alt text](images/emulation.png)



## Mitigation

For the mitigation of `Escape-to-host` we need nimbus-kyverno adapter to be in-place:
- First we need to install nimbus, you can do so by following the steps over [here](../../docs/getting-started.md#nimbus).
- Now you can follow the guide [here](../../docs/getting-started.md#nimbus-kyverno) to install nimbus-kyverno adapter.
- Now apply the escape-host-intent in your cluster as defined [here](../../examples/clusterscoped/escape-to-host-si-sib.yaml) and then try to re-run the attack, you'll see that now the agent will not be able to create a vulnerable pod. Resulting the failure in step-1 as defined above.