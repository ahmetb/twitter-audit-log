# Twitter audit trail backup

This repository backs up my
[follower list](followers.txt),
[following list](following.txt),
[blocked accounts list](blocked_accounts.txt) and
[muted accounts list](mutes.txt) periodically using GitHub Actions.

## Set up

> This code currently uses both Twitter v1 and v2 APIs. v2 API is currently
> behind a manual approval process.

1. Fork this repository.
1. `git rm *.txt` and commit.
1. Create a Twitter app.
1. Go to Repository Settings &rarr; Secrets and add secrets from your Twitter
   app:

   - TWITTER_CONSUMER_KEY
   - TWITTER_CONSUMER_SECRET
   - TWITTER_ACCESS_TOKEN
   - TWITTER_TOKEN_SECRET

1. See [.github/workflows/update.yml](/.github/workflows/update.yml) and modify
   the cron schedule (in UTC) as you see fit.

1. Commit and push. Once the time arrives, the cron would work, and commit
   the lists into `.txt` files and push the changes.
