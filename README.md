# Goloso

_Goloso is an Ansible callback server over NSQ_

The idea here it to have an automatic bootstrap mechanism using a callback system to run Ansible playbooks remotely.

Whenever a instance goes up, it posts a message saying “hi guys, I’m alive” to a central message broker. Them comes **Goloso**, our engine which consumes from a “bootstrap” topic and orders the playbook run into that fresh new and sweet instance.

This goloso points to a single Ansible playbook, the master entrypoint. This playbook knows what to choose based on the variables/tags received.

## Install Goloso

    go get github.com/gophergala/goloso

## Goloso Dependencies

    go get github.com/bitly/go-nsq
    go get github.com/boltdb/bolt

## Install NSQ

    brew install nsq
    
## Run NSQ

    nsqlookupd
    nsqd --lookupd-tcp-address=127.0.0.1:4160
    nsqadmin --lookupd-http-address=127.0.0.1:4161
    open http://localhost:4171

## Install Ansible

    brew install ansible

## Future

 * web admin interface
 * replay events