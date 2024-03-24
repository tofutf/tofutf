# Github

Configure tofutf to sign users in using their Github account.

Create an OAuth application in Github by following their [step-by-step instructions](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app).

* Set application name to something appropriate, e.g. `tofutf`
* Set the homepage URL to the URL of your tofutfd installation (although this is purely informational).
* Set an optional description.
* Set the authorization callback URL to:

    `https://<tofutf_hostname>/oauth/github/callback`

!!! note
    It is recommended that you first set the [`--hostname` flag](../../config/flags.md#-hostname) to a hostname that is accessible by Github, and that you use this hostname in the authorization callback URL above.

Once you've registered the application, note the client ID and secret.

Set the following flags when running `tofutfd`:

```
tofutfd --github-client-id=<client_id> --github-client-secret=<client_secret>
```

If you're using Github Enterprise you'll also need to inform `tofutfd` of its hostname:

```
tofutfd --github-hostname=<hostname>
```

Now when you start `tofutfd`, navigate to its URL in your browser and you'll be prompted to login with Github.

![github login button](../../images/github_login_button.png)

!!! note
    In previous versions of tofutf, Github organizations and teams were synchronised to tofutf. This functionality was removed as it was deemed a security risk.
