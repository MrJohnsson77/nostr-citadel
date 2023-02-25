# [Nostr Citadel - The Sovereign Relay | WiP](https://github.com/MrJohnsson77/nostr-citadel)  

Nostr Citadel, a personal relay that stores and safeguard all your nostr data.  

The idea is that anyone should be able to run their own relay and manage their own data in a **simple** and convenient way.
Once you're up and running, the relay will actively sync data from other relays and create a backup locally.
 
By default, only admin data is synced to the relay.  
Only admin and whitelisted npubs are allowed to post to the relay.

Stay Sovereign 🤙

## Disclaimer
This is me learning [Nostr](https://github.com/fiatjaf/nostr) and [Go](https://go.dev/).  
I will be adding all features that I want myself for a sweet relay.  

Let me know if there's anything specific you'd like to see implemented.

No prior Go experience so use it at your own peril. 💀  
It should be [safe](https://www.youtube.com/watch?v=dQw4w9WgXcQ) to use. 

As Is - No Warranty!

## Credits
Thanks to [fiatjaf](https://github.com/fiatjaf/relayer) for blueprint and inspiration. 💜

## Features / Todo

- [x] Nostr Relay
    * Core event model
- [x] Npub Whitelist
    * Whitelist your friends and foes for event posting
      * Limit reading in future version for private relays
- [x] Sync data to local relay
    * Sync either all whitelisted npubs or only admin data (events) to the relay
    * It will stay up to date by syncing once a day.
      * Using `since` so we don't pull everything all the time. 
      * Random interval so that not all citadels have the same sync interval
- [ ] Paid Relay
  * Core Lightning
  * LND
  * ...others...
- [ ] Export & Import Data
  * Export to file
  * Import
  * Bootstrap from export
- [ ] Cli Tool
- [ ] Local Dashboard
- [ ] Automatic SSL Termination
  * Lets Encrypt Certificate
- [ ] Citadel Admin & Nostr Client
  * Web - [Nostr Citadel Web Client](https://github.com/MrJohnsson77/nostr-citadel-watch) ( Coming "soon" )
    * Nostr Client & Admin Dashboard for Nostr Citadel
    * Standard client functionality
      * Global
      * Followers
      * Groups
      * Notifications
      * etc...
    * Follow recommendations
    * Trending on your relay ( and others... )
    * AI support, tweak and build your own recommendation and grouping engine
      * Export and share your models
- [ ] Remote Backup
- [ ] Additional database backends
  * Redis
  * PostgresSQL
  * MongoDB
  * MySQL
- [ ] Horizontal scaling
- [ ] Production Relay on wss://relay.nostr-citadel.io

  
## Nips

[NIPs](https://github.com/nostr-protocol/nips) with a relay-specific implementation are listed here.

- [x] NIP-01: [Basic protocol flow description](https://github.com/nostr-protocol/nips/blob/master/01.md)
- [ ] NIP-02: [Contact List and Petnames](https://github.com/nostr-protocol/nips/blob/master/02.md)
- [ ] NIP-03: [OpenTimestamps Attestations for Events](https://github.com/nostr-protocol/nips/blob/master/03.md)
- [ ] NIP-05: [Mapping Nostr keys to DNS-based internet identifiers](https://github.com/nostr-protocol/nips/blob/master/05.md)
- [x] NIP-09: [Event Deletion](https://github.com/nostr-protocol/nips/blob/master/09.md)
- [x] NIP-11: [Relay Information Document](https://github.com/nostr-protocol/nips/blob/master/11.md)
- [x] NIP-12: [Generic Tag Queries](https://github.com/nostr-protocol/nips/blob/master/12.md)
- [x] NIP-15: [End of Stored Events Notice](https://github.com/nostr-protocol/nips/blob/master/15.md)
- [x] NIP-16: [Event Treatment](https://github.com/nostr-protocol/nips/blob/master/16.md)
- [x] NIP-20: [Command Results](https://github.com/nostr-protocol/nips/blob/master/20.md)
- [ ] NIP-22: [Event `created_at` limits](https://github.com/nostr-protocol/nips/blob/master/22.md)
- [ ] NIP-26: [Event Delegation](https://github.com/nostr-protocol/nips/blob/master/26.md) 
- [ ] NIP-28: [Public Chat](https://github.com/nostr-protocol/nips/blob/master/28.md)
- [ ] NIP-33: [Parameterized Replaceable Events](https://github.com/nostr-protocol/nips/blob/master/33.md)
- [ ] NIP-42: [Authentication of clients to relays](https://github.com/nostr-protocol/nips/blob/master/42.md)


## Requirements
- Computer
- Internet
- Courage

## Get Started
Source and binaries will be available soon...   
Contact me if you want to do some alpha testing.

### Configuration & Operation
Add your npub and citadel_url in the config.yaml.

On first startup, the npub set as admin will be bootstrapped by downloading the profile data from several default
relays, when the profile is downloaded the relay will sync using the relays in the profile.

Initial startup will sync last 7 days, this can be tweaked in the config.

### Run from binary
Download binary for your architecture from the [releases](https://github.com/MrJohnsson77/nostr-citadel/releases) section.

### Run from source
  ```
  $ git clone git@github.com:MrJohnsson77/nostr-citadel.git
  $ cd nostr-citadel
  $ go run main --port 1337 
  ```

### Build from source
```
  $ git clone git@github.com:MrJohnsson77/nostr-citadel.git
  $ cd nostr-citadel
  $ make build
  ```

# Author

- [MrJohnsson](https://github.com/MrJohnsson77) - npub1fhdx6c6pt0ff6k3h5em760fzzzcehe9kqnjl05d2xwmg0ctjp80sn4hhsv 


# Contributors (A-Z)

- 

## License

This project is MIT licensed.
