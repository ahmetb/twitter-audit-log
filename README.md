# Twitter account backup/audit via GitHub Actions

This repository backs up various Twitter data of [my account](https://twitter.com/ahmetb):

- [follower list](followers.txt) (works up to 75k followers)
- [following list](following.txt)
- [blocked accounts list](blocked_accounts.txt) and
- [muted accounts list](mutes.txt) periodically using GitHub Actions.

(Twitter does not offer an API for exporting muted words.)

You can fork this repository and make it work for your account.

## Set up

> This code currently uses both Twitter v1 and v2 APIs. v2 API is currently
> behind a manual approval process.

1. Fork this repository.
1. `git rm *.txt` (delete my backups), `git commit`, `git push`
1. Create a Twitter app from [Developer Portal](https://developer.twitter.com/).
1. Go to GitHub `Repository Settings` &rarr; `Secrets` and add secrets from the
   Twitter app you created in previous step:

   - TWITTER_CONSUMER_KEY
   - TWITTER_CONSUMER_SECRET
   - TWITTER_ACCESS_TOKEN
   - TWITTER_TOKEN_SECRET

1. (Optional) Modify the cron schedule (in UTC) as you see fit in
   [.github/workflows/update.yml](/.github/workflows/update.yml).
   Commit and push.
   
1. GitHub will trigger the scheduled action and backup
   the lists to `.txt` files as commits to your forked repository.
