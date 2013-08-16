Character Encodings 101
---
As a new developer, you may be content floating in your wonderful little bubble of entirely-ASCII-encoded text, but sooner or later everyone has to interact with the real world, and the real world has an unfortunate variety of encodings.

This article will delve into, for some of the more important encodings, what they are, why they exist, and how they encode characters. In a follow-up article, I'll talk about how to handle these encodings properly in Ruby.

## US-ASCII

ASCII is perhaps the simplest and, until recently at least, most prevalent character encoding. It can encode 128 different characters, including upper and lower case of our Latin alphabet, the digits 0-9, as well as common punctuation and various control codes.

ASCII represents these 128 characters using 7 bits (because `log2(128)` is 7); however, computers don't deal well with values under 8 bits, so each ASCII character is stored in a full byte (8 bits). This means one bit is effectively unused in ASCII, and always set to 0. If this bit is not zero, the character is generally printed as an escaped numeric sequence (eg. binary 1110 0110 might print as '\xE6' in ASCII, while binary 0110 0110 prints as 'f', because the high bit is set in the former but not the latter).

ASCII is a very simple encoding to implement, which is why it is so prevalent, but it is also extremely limited. Many languages cannot be represented by the latin alphabet (even French requires various accents, not available in ASCII). Fortunately, ASCII only uses half of the possible values that can fit inside a byte. Thus began the proliferation of ASCII-compatible extended character sets.

## ISO/IEC 8859-1

Perhaps the most widely used extended-ASCII character encoding is 8859-1. Every ASCII value remains the same in this encoding, but the values with the high bit set map to actual characters, rather than unprintable values.

8859-1 is used to represent the additional characters required by Western European languages. For example, 'ü' is represented by binary 1111 1100. This encoding includes all manner of accented vowels, the German ß, several characters used mostly in Scandinavian languages, and a few other symbols.

## Windows Codepage 1252

Very similar to ISO-8859-1. Where 8859-1 has a range of bytes from `0x80` to `0x9F` that do not correspond to printable characters, codepage 1252 uses these bytes to code a handful of additional occasionally-useful symbols.

## ISO/IEC 8859-2

Essentially the same motivation as 8859-1, but for Eastern European (but still latin-based) languages. Think Czech, Polish.

## ISO/IEC 8859-\* and Windows-125\*

In fact, there are quite a number of these extended-ASCII encodings. 8859-1 for Western European, -2 for Eastern European, -3 for South European (Turkish, Maltese), -4 for North European (Estonian, Latvian, …), -5 for Cyrillic, -6 for Arabic, -7 for Greek, -8 for Hebrew, -9 for Turkish, -10 for Nordic languages, -11 for Thai, -13 for Baltic languages, -14 for Celtic, -15 is a variant of -1, and finally, -16 is for South-Eastern European languages.

Wow.

As if that weren't enough, most of these have Windows Codepage near-equivalents, where Microsoft imported their mappings, but in some cases, changed a few things around.

## Encoding bankruptcy

This is all a big mess. If you're given a byte sequence and told it's an extended-ASCII string, you really have no hope of figuring out what the high bits mean unless you speak about a dozen languages, or are specifically told the encoding.

The underlying problem is that one byte just isn't enough to represent a character. This is where Unicode comes in.

## Unicode

The Unicode standard came on the scene in the wake of this inane proliferation of single-byte encoding standards. While Unicode doesn't actually make any definition of how to encode a character into a sequence of bits or bytes, it does define one large, universally-standardized lookup table of characters. Each character is given what's called a "codepoint". You can think of a codepoint as an index into Unicode's sparse array of about a million possible characters.

Unicode's mission is not to define an Encoding standard, but to define a standard character set that can be represented using whatever encoding a programmer wishes to use. Several encodings for Unicode have been used over the years, the more relevant of which start with "UTF": "UTF-32", "UTF-16", and – most importantly – "UTF-8".

To drive this point home, take this character for example: ☃. Unicode Snowman. He is defined as "U+2603". That means that, however the character is encoded, it should encode the index as hexadecimal 2603.

## Can we make our own Unicode Encoding?

Sure! The maximum number of characters we could represent in Unicode is 1,114,112. We need 21 bits to do that, which rounds up to 3 bytes. We could simply encode each code point as a 3-byte integer. While this is very easy to conceptualize, there are a multitude of downsides to this scheme:

* The string will have NUL bytes in it. This is unfriendly to languages such as C that NUL-terminate strings.
* It's wasteful. Whereas ASCII takes 5 bytes to represent "hello", this encoding would take 15.
* If you start reading in the middle of a string, it's extremely difficult to figure out where one character ends and the next begins without backing up to the start.

## UTF-32

In actuality, the encoding described above is implemented as UTF-32, with a few changes. Want to know why no one uses UTF-32? Here's the string "hi" in ASCII:

```text
0110100001101001
```

And here it is in UTF-32:

```text
00000000000000001111111011111111
00000000000000000000000001101000
00000000000000000000000001101001
```

It's actually worse than our theoretical 3-byte encoding as far as space-efficiency goes, because it uses 4 bytes per character, with an additional 4 bytes at the start indicating how the rest of the string is encoded.

## The search for a better standard

After Unicode was introduced, a lot of "standards" came on the scene as candidate ways to represent Unicode text. UCS-2 and UTF-16 were the main ones for a while, but they both had severe limitations. 

Finally, Ken Thompson invented UTF-8 and all was good in the world.

## UTF-8

UTF-8 is the first sane encoding for Unicode.

Every ASCII character is represented in Unicode by its own byte value, and no byte that is not an ASCII character looks like one. In other words, UTF-8 maintains the same backward compatibility with ASCII that all the ISO/IEC 8859-* encodings do. This is much more than can be said for earlier Unicode encodings.

Additionally, ASCII text is by nature valid UTF-8 text. This has two benefits:

1. Re-encoding is not required to transform plain ASCII text to UTF-8 (unless it contains bytes whose high bits are set, in which case it's not *really* plain ASCII text.)
2. Space efficiency! UTF-32 requires four bytes to represent the letter 'a', but UTF-8 only requires one.

UTF-8 represents longer Unicode codepoints by using multibyte characters. While this does mean that the byte-length of a character varies from 1-4 bytes, it is implemented in a clever way:

1. Every byte of a multibyte character has the high bit set.
2. The leading byte of a multibyte character has the length of the multibyte character encoded in a simple way.

If a multibyte character is two bytes long, the leading byte starts with "110". The number of 1s indicates the length of the character. If it starts with "1110", it is 3 bytes long, and if it starts with "11110", it is four bytes long. Each continuation byte begins with "10". The remaining bits in each of these bytes are populated by bits of the unicode codepoint.

In a more visual format:

```text
110x xxxx  10xx xxxx
[ 2 byte character ]

1110 xxxx  10xx xxxx  10xx xxxx
[ 3 byte character            ]

1111 0xxx  10xx xxxx  10xx xxxx  10xx xxxx
[ 4 byte character                       ]
```

In each of these examples, the character's length in bytes is indicated by the number of 1s preceding the first zero in the leading byte. Each successive byte of the multibyte character begins with "10". Each byte marked with an "x" is filled with bits from the actual unicode codepoint. For example, to encode the Unicode Snowman (U+2603):

```text
0x2603 = 0010 0110 0000 0011
```

This is 14 bits. There are only 11 'x's in the 2-byte character, so we need to use a 3-byte character to fit all these bits into the available 'x's.

```text
1110 xxxx  10xx xxxx  10xx xxxx
     0010    01 1000    00 0011
```

This gives us `1110 0010 1001 1000 1000 0011`, or `0xE29883` – the UTF-8 representation for the Snowman (☃).

Subtle-but-important points about the UTF-8 format:

* You can treat it as ASCII for a lot of puproses. If you are searching for tabs (`0x09`), stripping out angle-brackets (`0x3C` and `0x3E`), or similar, you can completely ignore the fact that you are working on a UTF-8 string. This is because no byte in a multibyte character ever looks like ASCII, and all ASCII encodings are the same in UTF-8.
* calculating string length is more difficult, but not bad, since multibyte characters encode their length in the first byte. You can simply jump ahead to the start of the next character after reading the leading byte.
* If you start reading in the middle of a  string, you never need back up of skip ahead by more than three bytes to synchronize to the beginning of a character, and and you know you are mid-character if the current byte begins "10". This is much trickier in essentially every other Unicode encoding.

## Conclusions

This article has described ASCII, various extended-ASCII encodings, glossed over serveral early Unicode encodings, and described UTF-8 in some detail. Hopefully it has provided some useful knowledge of the interaction between ASCII and UTF-8 character strings, and some motivation for the "always use UTF-8" mantra that gets repeated so often.

A future article will explain how to use and take advantage of these encodings in Ruby, and perhaps Go.

