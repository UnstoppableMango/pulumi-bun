# Experimental Pulumi Bun Support

Just me fiddling around and seeing if I can implement a Pulumi language provider for Bun now that it has HTTP2 server support.

## References

<https://www.pulumi.com/docs/iac/support/faq/#how-can-i-add-support-for-my-favorite-language>

<https://github.com/pulumi/pulumi/wiki/New-Language-Bring-up>

<https://github.com/pulumi/pulumi/tree/master/sdk/nodejs/cmd/pulumi-language-nodejs>

<https://github.com/pulumi/pulumi-dotnet/blob/main/pulumi-language-dotnet/main.go>

<https://github.com/pulumi/pulumi/pull/1456>

## Notes

Note to future me because I will inevitably forget, Renovate supports tracking git submodules by tag but git isn't a huge fan of this.

<https://docs.renovatebot.com/modules/manager/git-submodules/#updating-to-specific-tag-values>

TL;DR

> **Note:** Using this approach will disrupt the native git submodule update experience when using git submodule update --remote. You may encounter an error like fatal: Unable to find refs/remotes/origin/v0.0.1 revision in submodule path... because Git can only update submodules when tracking a branch. To manually update the submodule, navigate to the submodule directory and run the following commands: `git fetch && git checkout <new tag>`.
