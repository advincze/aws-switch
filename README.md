# aws-switch

switch aws profiles for the command line

aws-switch copies the specified profile to the default 
(and archives the current default if it is not found in the other profiles)

#### install:

```
go get -u github.com/advincze/aws-switch
```

#### use: 

```
$ aws-switch [profile]
```

profile usage suffix search, so partial matches are also found if they are not ambigous

