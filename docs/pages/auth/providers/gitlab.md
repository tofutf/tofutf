# Gitlab

Configure tofutf to sign users in using their Gitlab account.

Create an OAuth application for your Gitlab group by following their [step-by-step instructions](https://docs.gitlab.com/ee/integration/oauth_provider.html#group-owned-applications).

* Set name to something appropriate, e.g. `tofutf`
* Select `Confidential`.
* Select the `read_api` and `read_user` scopes.
* Set the redirect URI to:

    `https://<tofutfd_install_hostname>/oauth/gitlab/callback`

!!! note
    It is recommended that you first set the [`--hostname` flag](../../config/flags.md#-hostname) to a hostname that is accessible by Gitlab, and that you use this hostname in the redirect URI above.

Once you've created the application, note the Application ID and Secret.

Set the following flags when running `tofutfd`:

```
tofutfd --gitlab-client-id=<application_id> --gitlab-client-secret=<secret>
```

If you're hosting your own Gitlab you'll also need to inform `tofutfd` of its hostname:

```
tofutfd --gitlab-hostname=<hostname>
```

Now when you start `tofutfd` navigate to its URL in your browser and you'll be prompted to login with Gitlab.

!!! note
    In previous versions of tofutf, Gitlab groups were synchronised to tofutf. This functionality was removed as it was deemed a security risk.
