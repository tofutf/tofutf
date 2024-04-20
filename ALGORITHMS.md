# Architecture

TofutTF's architecture is fairly simple. There are several key components.

* Storage
* API
* Web
* Execution

### Storage

TofuTF stores all data inside of a single postgres database. This database stores both normal tabular data, in addition to binary data. Storing binary data inside of a sql database may not scale as well as with a dedicated external storage system like s3, however it allows TofuTF to maintain a small footprint in your infrastructure.

## API

TofuTF's api is primarily an implementation of the Terraform Cloud api. This is partly due to TofuTF's nature as a TFC self-hosted reimplementation. This is what allows terraform, and opentofu to both interact with tofutf. This also has the added benefit that the [TFE (Terraform Enterprise/Terraform Cloud)](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs) provider can be used to configure TofuTF.

There are several endpoints that are exclusive to TofuTF such as the `/versions` endpoint, however in general `TofuTF` strives to have a minimal api surface area. In the future, if the need arises for dedicated api endpoints, they will most likely be accompanied with the introduction of a TofuTF provider.

## Web

TofuTF uses HTMX, and server side rendered templating to serve the web interface. HTMX is used to enrich certain views, and make them more dynamic, but most of TofuTF functions without it.

## Execution

Notes on various concepts in use by TFE/TFC/otf.

### Run Status/State

https://www.terraform.io/cloud-docs/api-docs/run#run-states

A run begins life in the `pending` state. The next state (in otf) is `plan_queued`, indicating it's ready to enter the plan phase and that it's currently waiting in the global queue (see below). A run is switched from `pending` to `plan_queued` if:

* it's reached the front of the workspace queue (see below)
* it's a speculative run (i.e. plan-only) in which case the switch occurs immediately

### Workspace Queue

Each workspace maintains a queue of runs. The run at the head of the queue is the 'current' run for that workspace, i.e. the run currently, or shortly, to be executed by an agent. It blocks any runs behind it in the queue. Only once it has reached a completed state it's removed from the queue and the next run takes its place.

Note: speculative runs (i.e. plan-only runs) are not queued.

### Global Queue

https://www.terraform.io/cloud-docs/run/run-environment

The global queue is a queue of run phases awaiting execution by an agent. According to the referenced document, the following priorities are applied:

1. Applies that will make changes to infrastructure have the highest priority.
2. Normal plans have the next highest priority.
3. Speculative plans have the lowest priority.

Note: the workspace queue is a queue of *runs* whereas the global queue is a queue of *run phases* i.e. plans and applies. In the former the entire run needs to enter a completed state before it is removed which may entail a plan followed by an apply. Whereas in the latter case only the run phase need have entered a completed state before it is removed.

### Agents

Note: this only applies to otf.

Agents execute run phases i.e. plans and applies.

By default otfd has 'local' agent embedded in the `otfd` binary. The local agent executes run phases belonging to *any* organization.

Optionally, 'remote' agents can be deployed. They connect to otfd over HTTPS. They authenticate using a token which is created via the Web UI. The token is scoped to an organization, permitting the agent to execute phases belonging to runs in that organization and that organization only.

Agents can run multiple run phases concurrently. (Note this is different to the TFC agent which executes only one at a time).
