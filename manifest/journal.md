# Journal

A descritive documentation about decisions make on this project.

## 2020-11-10 Morning decisions

### Max pool size 2³²-1

I was wondering if I should use an int64 on the pool size but the true is if you have more than 2³²-1 queued on you pool, the machine should not have enouth memory to handle it.

That is why I will use a simple int32 on it. (I'm talking about 4294967295 queued, dude, please upscale your application )

### Metrics

It would be nice have some metrics about what is happening on the rotines pool, I will put a minimal to make this project run and later I will try to add some fancy metrics.

## 2020-11-13 using golang standard layout

Golang standart layout is not that great layout but it is still something.

## 2020-11-21 Queueds Runnins and Linked List

I think there is a oportunity to improve about how the workers are queueds
it may have problems with concurrency unexpected because it was poor made
because it was not made expecting add new workers after it was started.
