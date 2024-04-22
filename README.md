<h1 align="center"> TofuTF </h1>
<p align="center">
    <img src="readme_logo.png" />
</p>

![Build](https://github.com/tofutf/tofutf/actions/workflows/build.yml/badge.svg)  ![GitHub License](https://img.shields.io/github/license/tofutf/tofutf) ![GitHub Release](https://img.shields.io/github/v/release/tofutf/tofutf) [![Star on GitHub](https://img.shields.io/github/stars/tofutf/tofutf.svg?style=flat)](https://github.com/tofutf/tofutf/stargazers) ![GitHub contributors from allcontributors.org](https://img.shields.io/github/all-contributors/tofutf/tofutf) ![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=flat&logo=postgresql&logoColor=white) ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) ![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=flat&logo=kubernetes&logoColor=white) ![Terraform](https://img.shields.io/badge/terraform-%235835CC.svg?style=flat&logo=terraform&logoColor=white) ![Maintained-Yes](https://img.shields.io/badge/Maintained%3F-yes-green.svg?style=flat) [![OpenSSF Best Practices](https://www.bestpractices.dev/projects/8815/badge)](https://www.bestpractices.dev/projects/8815)[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/tofutf/tofutf/badge)](https://scorecard.dev/viewer/?uri=github.com/tofutf/tofutf)

TofuTF is an open source alternative to Terraform Enterprise. Includes SSO, team management, agents, no per-resource pricing, and soon to be [OpenTofu](https://opentofu.org/) support.

## Getting Started

### Quick Start

Create a file named `values.yaml` and paste the following contents inside of it.

```yaml
# values.yaml

# The secret is used to sign sessions. It should be kept confidential, and 
# production installs of tofutf should have a randomly generated secret.
secret: 2876cb147697052eec5b3cdb56211681

# The siteToken is the special token that grants administrator access to 
# tofutf. Production installs of tofutf should have a randomly generated
# site token.
siteToken: site-token

# here we enable the bundled postgres instance, and configure it to provision
# a tofutf database.
postgres:
  enabled: true
  database: tofutf

# here we configure tofutf to connect to the bundled postgres instance. 
database: postgres://tofutf-postgresql/tofutf?user=postgres
databasePasswordFromSecret:
  name: tofutf-postgresql
  key: postgres-password
```

Then run the following command to install tofutf.

```
helm install my-release -f values.yaml oci://ghcr.io/tofutf/tofutf/charts/tofutf --version v0.8.0
```

### Congrats! 🎉
Congrats, you have deployed TofuTF! Check the quickstart guide on the official docs site for next steps. https://docs.tofutf.io/quickstart

## Legal

TofuTF is in no way affiliated with Hashicorp. Terraform and Terraform Enterprise are trademarks of Hashicorp. Hashicorp have [confirmed](https://www.reddit.com/r/Terraform/comments/15p2p32/impact_of_new_licensing_on_open_source/) TofuTF is in compliance with their BSL license.

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/leg100"><img src="https://avatars.githubusercontent.com/u/75728?v=4?s=32" width="32px;" alt="Louis Garman"/><br /><sub><b>Louis Garman</b></sub></a><br /><a href="https://github.com/tofutf/tofutf/commits?author=leg100" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="http://blog.johnrowley.co"><img src="https://avatars.githubusercontent.com/u/3454480?v=4?s=32" width="32px;" alt="John Rowley"/><br /><sub><b>John Rowley</b></sub></a><br /><a href="https://github.com/tofutf/tofutf/commits?author=robbert229" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/jpetrucciani"><img src="https://avatars.githubusercontent.com/u/8117202?v=4?s=32" width="32px;" alt="jacobi petrucciani"/><br /><sub><b>jacobi petrucciani</b></sub></a><br /><a href="https://github.com/tofutf/tofutf/commits?author=jpetrucciani" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/m-guesnon-pvotal"><img src="https://avatars.githubusercontent.com/u/91205142?v=4?s=32" width="32px;" alt="m-guesnon-pvotal"/><br /><sub><b>m-guesnon-pvotal</b></sub></a><br /><a href="https://github.com/tofutf/tofutf/commits?author=m-guesnon-pvotal" title="Code">💻</a></td>
    </tr>
  </tbody>
  <tfoot>
    <tr>
      <td align="center" size="13px" colspan="7">
        <img src="https://raw.githubusercontent.com/all-contributors/all-contributors-cli/1b8533af435da9854653492b1327a23a4dbd0a10/assets/logo-small.svg">
          <a href="https://all-contributors.js.org/docs/en/bot/usage">Add your contributions</a>
        </img>
      </td>
    </tr>
  </tfoot>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## Provenance

TofuTF is a fork of the now abandoned [otf](https://github.com/leg100/otf). Louis Garman did some amazing work, and this fork is an attempt to carry the torch.

<img src="readme_otf_logo.png" width="128px"/>
