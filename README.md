# Git Auth Proxy

Proxy to allow single-user control of locally managed GitHub credentials.

Most Git providers offer multiple ways of authenticating when cloning repositories and communicating with their API. These authentication methods are usually tied to a specific user and in the best
case offer the ability to scope the permissions. The lack of organization API keys leads to solutions like GitHubs solution to [create a machine user](https://docs.github.com/en/developers/overview/managing-deploy-keys#machine-users)
that has limited permissions. The need for machine user accounts is especially important for GitOps deployment flows with projects like [Flux](https://docs.github.com/en/developers/overview/managing-deploy-keys#machine-users)
and [ArgoCD](https://github.com/argoproj/argo-cd). These tools need an authentication method that supports accessing multiple repositories, without sharing the global credentials with all users.

<p align="center">
  <img src="./assets/architecture.png">
</p>

Git Auth Proxy attempts to solve this problem by implementing its own authentication and authorization layer in between the client and the Git provider. It works by using static tokens that are
specific to a Git repository. These tokens are then used by users when interacting with the remote through the proxy. When a repository is cloned through the
proxy, the token will be checked against the list of policies. If a match is found for the token and the repository cloned, it will be replaced with the correct credentials in the downstream connection. The request will be denied if a token is used to clone any other
repository which is does not have access to.

## How To

The proxy reads its configuration from a JSON file. It contains a list of repositories that can be accessed through the proxy.

When using GitHub a GitHub Access Token is used.

```json
{
  "policies": [
    {
      "provider": "github",
      "github": {
        "token": "<ACTUAL_GITHUB_TOKEN>"
      },
      "userAuth": {
        "tokenHash": "<HASH_OF_USER_TOKEN>"
      },
      "host": "github.com",
      "repositories": [
        {
          "owner": "yourorg",
          "name": "fleet-infra"
        }
      ]
    }
  ]
}
```

### Git

Cloning a repository through the proxy is not too different from doing so directly from GitHub. The only limitation is that it is not possible to clone through ssh, as Git Auth Proxy
only proxies HTTP(S) traffic. To clone the repository `repo-1` get the clone URL from the repository page.
Then replace the host part of the URL with `git-auth-proxy` and add the token as a basic auth parameter. The result should be similar to below.

```shell
git clone http://<token-1>@git-auth-proxy/org/_git/repo-1
```

### API

API calls can also be done through the proxy. Currently only repository specific requests will be permitted as authorization is done per repository. This may change in future releases.

#### GitHub

The proxy assumes that the requests sent to it are in a GitHub enterprise format due to the way GitHub clients behave when configured with a host that is not `github.com`. The main difference between
GitHub Enterprise and non GitHub Enterprise is the API format. The GitHub Enterprise API expects all requests to the API to have the prefix `/api/v3/` while non GitHub Enterprise API requests are sent
to the host `api.github.com`.

# License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

