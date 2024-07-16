Local Environment
-----------------

To be able to run the tests against a real Gitlab instance, just run.

```shell
bash initial-setup.sh
```

This should setup a Gitlab instance that is fully configured for the tests locally. 

As configuring takes quite a bit of time. After the first start you can run the command bellow.

```shell
bash backup-volumes.sh
```

And to restore it back to the original setting 

```shell
bash restore-volumes.sh
```