# serving-cert-keystore-operator
An Operator which extends "service.alpha.openshift.io/serving-cert-secret-name" annotation to to create a keystore

# Example

```yaml
kind: Service
apiVersion: v1
metadata:
  name: nginx-ex
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: nginx-ex-tls
    ykoer.github.com/serving-cert-create-pkcs12: 'true'
spec:
  ...
```

Add annotation via command line
```
$ oc annotate service nginx-ex 'ykoer.github.com/serving-cert-create-pkcs12=true' --overwrite
```

Remove annotation via command line
```
$ oc annotate service nginx-ex ykoer.github.com/serving-cert-create-pkcs12-
```

The Updated Secret will look like:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: nginx-ex-tls
data:
  tls.crt: >-
    ***
  tls.key: >-
    ***
  tls.p12: >-
    ***
  tls-pkcs12-password: ****
    
type: kubernetes.io/tls
```
