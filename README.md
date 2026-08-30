**`pulse`** is the CLI tool we'll build together: a **concurrent service health-checker**.
You give it a list of services (URLs), it pings them all *in parallel*, and reports which are up,
which are down, and how slow they are — with config files, subcommands, tests, and a release binary.
