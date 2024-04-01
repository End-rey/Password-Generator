# Password Generator

Generates a password from four words from the English dictionary, such that the distance between letters on the keyboard is minimized. The password length should be between 20 and 24.

## Data

In folder [`data`](/data) there is a file [`linux_words`](/data/linux_words) from `/usr/share/dict/words`. Also there is file [`words.txt`](/data/words.txt) and [`words_alpha.txt`](/data/words_alpha.txt) from this [repository](https://github.com/dwyl/english-words).
The file contains a list of words, one per line.

## Solution

I use words with minimum distance between letters on the keyboard, calculating like difference between length of word and distance between letters.

The main idea that I use map, where the key is the word and the value is a slice of words such that the distance between their first letters and the last letter of the key is less than `distanceBetweenWords`. This map optimizes the four nested loops. 
Also I use goroutines to make parallel execution.

## Output:

```cmd
43695 words loaded
34 words with difference <= -1 between distance and length
Distance between words < 2
Min distance:  12
Total passwords With Min distance:  62
Random password with min distance:  wedded dded deed deeded
Total passwords:  2723
Random password:  free wedder freer treed
```

I calculate a passwords with min distance that possible, and the length of this passwords are 20.
And then take a random password with min distance.

Also I calculate passwords, where words with minimum distance, and length of passwords is between 20 and 24. 
And then take a random password from all of them.