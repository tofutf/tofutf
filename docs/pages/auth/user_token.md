# User Tokens

A user can generate API tokens. The token shares the same permissions as the user.

To manage your tokens, go to **Profile > Tokens**.

![profile page](../images/user_tokens.png){.screenshot .crop}
![new token enter description](../images/user_token_enter_description.png){.screenshot .crop}
![user token created](../images/user_token_created.png){.screenshot .crop}

API tokens are not only used for programmatic access but for authenticating `terraform` and `tofutf`. For example, you can use `terraform login` to store a token on your workstation:

```bash
terraform login <tofutf hostname>
```

And follow the instructions. The token is persisted to a local credentials file for use by both `terraform` and `tofutf`.
