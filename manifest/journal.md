# Journal

A descritive documentation about decisions make on this project.

## 2020-11-10 Morning decisions

### Max pool size 2³²-1

I was wondering if I should use an int64 on the pool size but the true is if you have more than 2³²-1 queued on you pool, the machine should not have enouth memory to handle it.

That is why I will use a simple int32 on it. (I'm talking about 4294967295 queued, if you need more tasks, please upscale your application )

### Metrics

It would be nice have some metrics about what is happening on the rotines pool, I will put a minimal to make this project run and later I will try to add some fancy metrics.

## 2020-11-13 using golang standard layout

Golang standart layout is not that great layout but it is still something.

## 2020-11-21 Queueds Runnins and Linked List

I think there is a oportunity to improve about how the workers are queueds
it may have problems with concurrency unexpected because it was poor made
because it was not made expecting add new workers after it was started.

## 2020-12-13 When should I send information about task done?

There is a lot of places that I can use like:

* On worker after the given function finished to run. I don't like this aprouch because I think it is from pool domain, and I will just be adding more variables on queued jobs.

* On same go rotine that listen done task at first before call other go rotine from queue. I liked the ideia of putting on the for of chDone, but I will put it on the end, because the priority is on running the pool, and send information about what is happening will have less priority, but still will be send anyway.

## 2020-12-13 I is lock time

The tests get some race conditions situations, to fix it I will put mutation locks where it was
happening.

I hope those race condition will be fixed.

It is not really a big problem, I started to do this pool to learn more about multi-thread strategies any way.


## 2020-12-16 Cloase channels with channels ...

It is a little strange for me, but it seams to be the best solution to close channels on range
without happening race conditions.

I took a look into some methods and I think that I will go with the selection option

select:
case <- chShutdown
break

something like that
