# siasn-libs-backend

This is a general repository that hosts many reusable packages that can be used to create
SIASN backend services.

**siasn-libs-backend use feature branch model, not gitflow, so development branch is not needed.**

## config

The `config` package holds a config related functions used to get configuration from environment
variables. Since SIASN services are all twelve factor apps, we have dropped support for
config files and use environment variables completely. We have also dropped the support for
changing configurations in runtime, which often requires the creation of a server to accept
request to change configurations.

The package contains mostly getter functions to fetch and parse environment variables.

### Using `config` Package

To use `config` package, create a new `ServiceConfig` and a struct to store your configuration
for a particular service/package/feature that you implement.

```go
import (
    "github.com/fazrithe/siasn-jf-backend-git/libs/config"
)

type MyConfig struct {
    Port int `config:"PORT"`
    Host string `config:"HOST"`
}

sc := config.ServiceConfig{
    Prefix: "MYSERVICE",
    ArraySeparator: " ",
}
```

Notice that we have `config` struct tags set to each of `MyConfig` field. This way we can
use `ServiceConfig.ParseTo` method to automatically parse the environment variables for us, and
fill all our struct fields. The `Port` field value will be taken from `MYSERVICE_PORT` environment
variable, and the `Host` will be taken from `MYSERVICE_HOST` environment variable.

If an environment variable for a particular field does not exist, the field is simply skipped, and the
initial value in the field is left intact. This way you can give a prefilled struct that already have some
fields filled with default values in case some environment variables were not supplied by the user.

If you need only a few configurations, maybe less than three, you can choose to use the individual
getters, which will return `ErrConfigNotFound` if the variable can not be found if using the getter
without default values.

### Future Improvements of `config` Package

Adding some more new getters, like slice of integers, `net.IP`, and so on.

## Legal and Acknowledgements

This repository was built by:
* Sergio Ryan \[[sergioryan@potatobeans.id](mailto:sergioryan@potatobeans.id)]
* Eka Novendra \[[novendraw@potatobeans.id](mailto:novendraw@potatobeans.id)]
* Stefanus Ardi Mulia \[[stefanusardi@potatobeans.id](mailto:stefanusardi@potatobeans.id)]
* Audie Masola Putra \[[audiemasola@potatobeans.id](mailto:audiemasola@potatobeans.id)]

Copyright &copy; 2021 Indonesian National Civil Service Agency (Badan Kepegawaian Negara, BKN).  
All rights reserved.