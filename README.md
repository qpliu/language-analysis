Code to download transcripts and analyze the language used in them.

Motivation
----------
When listening to the radio, I noticed that at the end of a segment,
the host often thanks the guest or correspondent, who then usually
responds with something like "thank you" or "thank you for having me".
Occasionally, the response is "you're welcome" or "my pleasure" or
something in that vein.  But sometimes, the response is "you bet".
Once I thought about it, I don't know what "you bet" means or what
its origin is.

I had the idea of having code analyze transcripts of radio shows to
see what the distribution of these final responses are and how that
distribution changes over time.  I did nothing with this idea for
almost a decade, but it stayed in the back of my mind.

Now, I start writing the code to do it.

Analyses
========
```thank-collect```
-------------------
```thank-collect``` looks at up to 20 words in the final response
of a speaker in a transcript immediately following "thank" or "thanks",
and tabulates every unique group of 1 to 5 consecutive words in
that response by transcript and speaker.  Since each transcript has
a date, it should be possible to see how the relative prevalences of
groups of words change over time.

```phrase-collect```
--------------------
```phrase-collect``` tabulates the number of occurrences of a
specified set of phrases in each transcript, as well as the number
of occurrences of another set of phrases that preface a response
in each transcript.

Two of the phrases are "bucket list" and "perfect storm", since I
believe they should not appear until after the movies of the same
name were released.

I'm also interested in the relative prevelances of "definitely" and
"absolutely", since I believe the latter has become more prevelant
and the former has become less prevelant.

I'll also look at "exponentially", though I believe the growth in
the usage of that term plateaued before 2004, and the transcripts
I'm looking at only go back to 2004.

Also, I'm interested in responses prefaced by "look" and prefaced
by "absolutely".

Initial results
---------------
For 911 transcripts from between 2025-11-01 and 2025-11-30, the
most common final responses to thanks, with a lot of overlap, are

|response           |count|
|-------------------|-----|
|thank you          |285  |
|you're welcome     |67   |
|for having         |57   |
|for having me      |52   |
|so much            |28   |
|very much          |10   |
|you bet            |10   |
|appreciate it      |8    |
|a pleasure         |8    |
