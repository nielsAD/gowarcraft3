Explanation of W3M and W3X Files
================================

```
Author: Chocobo
Source: http://world-editor-tutorials.thehelper.net/cat_usersubmit.php?view=42787
```

1.01 Warcraft III Environment
-----------------------------

Most of computers have Warcraft III Installed on the path "C:\Program Files\Warcraft III\", which contains some files.

There are in the directory :
- war3.mpq (This is the RoC Environment used by Warcraft III)
- war3x.mpq (This is the TFT Environment)
- war3xlocal.mpq (This is the Local Files Environment if you have TFT)
- war3patch.mpq (This is the Patch Environment)
- (4)Lost Temple.w3m (An example of RoC map, this is from your Warcraft 3 Path + \Maps\\, and you will find a lot of other RoC maps)
- (2)Circumvention.w3x (An example of TFT map, located at "\Maps\FrozenThrone\\" if you have TFT)
- DemoCampaign.w3n (A demo campaign if you have TFT, located with the path Warcraft 3 Path + "\Campaigns\\"

1.02 Files Name in the Directory
--------------------------------

There are 4 differants tags names in the list at Chapter 1.0.
- .mpq
- .w3m
- .w3x
- .w3n

MPQ Files are files that contains files with their own directory inside the MPQ, without creating a directory in the main structure (C:\, D:\...). They are like .rar and .zip files.
w3m Files are RoC Maps. They are playable in RoC and in TFT.
w3x Files are TFT Maps. They are only playable in TFT.
w3n Files are Custom Campaign Maps. Again, they are only playable in TFT.


1.03 Warcraft III Data Format
-----------------------------

Warcraft III Data Format has 8 differants format type.

- Integers
Integers are 4 bytes in Little Endian Order
Example : 1234 Not Equal [00 00 04 D2]h
1234 Equal [D2 04 00 00]h

- Small Integers
Small integers are from -16384 to 16383. They take 2 bytes and are in Little Endian Order.

- Reals
Reals are Floats. They take 4 bytes and are in Little Endian Order.
Example : 7654.32 ~ [8F 32 EF 45]h ~ 7654.319824
The last number is the closest number to 7654.32

- Arrays
1 Array take 1 byte

- UnitIds
UnitId Integers takes 32 bytes. (46656 possibilities, they are like normal Integers)

- Flags
Boolean, or "Flags", take 4 bytes. In those 4 bytes, there are 32 bit, which can contains 32 flags. Each Flag can contain only the value 0 (False) or 1 (True).

- Water
Water Level (In the terrain) takes 4 bytes every cell, it handles 2 bytes for a total of 16 flags, and the 2 lasts for the closest number to the Water Level.

- Custom Handles
An integer or a flag may share themselves some bytes. A byte may handle two or more differant data.

- Structured Handles
Unknown. They have various size.

- Strings / Trigger Strings
Strings are just like arrays of chars finished with a null char (From C++ : "\0").
But, Blizzard uses a special coloring code to change the color of the text shown, these starts with "|c00", and finishes with "|r", most of the times. An example is "|c00BBGGRR|r", in which you replace by the percentage value "BB", "GG", "RR", by the percentage value you want. They are hexadecimal values (0123456789ABCDEF), using 2 digits each, BB = Blue, GG = Green, RR = Red.
In Trigger Strings, if it starts with "TRIGSTR\_", with a sensitive case, it is a Trigger String. Trigger Strings are ever keep in the virtual memory when you play a map, which is loaded at the map initalization, as a name of "TRIGSTR\_\*\*\*". Instead of writing the TRIGSTR\_<WHATEVER> thing, Warcraft III get a look in the string table, and displays the correct trigger string. They work only for files inside a .w3m or .w3x map, but not with the exception of WTS files, which is used for Trigger Strings files itself. If the following number after "TRIGSTR\_" is negative, it will be considered as the first one, called "TRIGSTR\_000". And when there are letters, they will be considered as 0, or if an example of "5aa", as 5.

At all, they take :
- Strings : (String Length + 1) bytes (Because of the finishing char)
- Trigger Strings : 12 bytes (Note Trigger Strings handles Strings, so the total is Trigger Strings + Strings)


1.10 .w3m and .w3x Files Format
-------------------------------

.w3m and .w3x files are Warcraft III Scenario Maps, which are like MPQ Files. They both takes a 512 bytes header format, but for some authentification, .w3x files takes an extra 260 bytes in the header.

Here is the header file of .w3m files :
```
char[4]: "HM3W"
int: unknown
string: map name
int: map flags
0x0001: if 1=hide minimap in preview screens
0x0002: if 1=modify ally priorities
0x0004: if 1=melee map
0x0008: if 1=playable map size was large and has never been reduced to medium
0x0010: if 1=masked area are partially visible
0x0020: if 1=fixed player setting for custom forces
0x0040: if 1=use custom forces
0x0080: if 1=use custom techtree
0x0100: if 1=use custom abilities
0x0200: if 1=use custom upgrades
0x0400: if 1=map properties menu opened at least once since map creation
0x0800: if 1=show water waves on cliff shores
0x1000: if 1=show water waves on rolling shores
int: max number of players
```

And this is for .w3x :

```
char[4]: "NGIS"
byte[256]: the 256 bytes for authentification
```

Inside a .w3m and a .w3x file, you may find those files :

```
(signature)
(attributes)
war3map.w3e
war3map.w3i
war3map.wtg
war3map.wct
war3map.wts
war3map.j
war3map.shd
war3mapMap.blp
war3mapMap.b00
war3mapMap.tga
war3mapPreview.tga
war3map.mmp
war3mapPath.tga
war3map.wpm
war3map.doo
war3mapUnits.doo
war3map.w3r
war3map.w3c
war3map.w3s
war3map.w3u
war3map.w3t
war3map.w3a
war3map.w3b
war3map.w3d
war3map.w3q
war3mapMisc.txt
war3mapSkin.txt
war3mapExtra.txt
war3map.imp
```


1.11 The war3map.j file : JASS2 Script
--------------------------------------

This is the main map script file. It's a text file and you can open it with notepad.
Sometimes it's renamed to Scripts\war3map.j by map protectors to keep you away from it.
The language used is called JASS2 and has been developed by Blizzard. It's a case sensitive language.
When you play a map, the jass script is loaded and executed.
When you select a map in when creating a game Warcraft III will first look up the "config" function and execute its code to set up the player slots.
Then, when the game has started, Warcraft III looks for the function called "main" and executes it.
You may find more informations at : http://jass.sourceforge.net/doc/


1.12 The war3map.w3e file : The Environment
-------------------------------------------

This is the tileset file. It contains all the data about the tilesets of the map.
The map is divided into squares, which contains tiles, and which has 4 corners. Each map size would be so for an example, 257x257 instead of 256x256.

Here is the file format :

```
A  	 Ashenvale
B  	Barrens
C 	Felwood
D 	Dungeon
F  	Lordaeron Fall
G 	Underground
L 	Lordaeron Summer
N 	Northrend
Q 	Village Fall
V 	Village
W 	Lordaeron Winter
X 	Dalaran
Y 	Cityscape
Z 	Sunken Ruins
I 	Icecrown
J 	Dalaran Ruins
O 	Outland
K 	Black Citadel
```

And this is the header :

```
char[4]: "W3E!"
int: w3e format version [0B 00 00 00]h = version 11
char: main tileset [TS]
```

I may explain a lot about this file later.


1.13 The war3map.shd file : The Shadow Map File
-----------------------------------------------

This file has no header, only raw data.

```
Size of the file = 16*map_width*map_height
1 byte can have 2 values:
00h = no shadow
FFh = shadow
Each byte set the shadow status of 1/16 of a tileset.
It means that each tileset is divided in 16 parts (4*4).
```


1.14 war3mapPath.tga The Image Path File and/or "war3map.wpm" The Path Map File
-------------------------------------------------------------------------------

Only one of these two file is used for pathing. Old Warcraft 3 beta versions lesser or equal to 1.21 uses the "war3mapPath.tga" file.
Since beta 1.30, Warcraft 3 uses a new file format instead: "war3map.wpm".


1.15 The war3mapPath.tga file : The Image Path File
---------------------------------------------------

It's an standard 32bits RGB TGA file with no compression and a black alpha channel. The TGA format is really important because if Warcraft III doesn't recognise the file format, it'll do weird things on the tilesets (like put blight everywhere)! Don't forget the alpha channel! Each tile of the map is divided in 16 pixels (4\*4 like in the shadow file), so the TGA width is 4\*map\_width and its height is 4\*map\_height and each pixel on the TGA affects a particular part of a tileset on the map. The color of a pixel sets the rules for that part. The top left corner of the image is the upper left corner on the map.

Header format (18 bytes) :

```
byte: ID Length = 0
byte: Color Map Type = 0
byte: Image Type = 2 (uncompressed RGB)
-- Color Map Specification (5 bytes) --
byte[2]: First Entry Index = 0
byte[2]: Color Map Length = 0
byte: Color Map Entry Size = 0
-- Image Spec (10 bytes) --
byte[2]: X origin = 0
byte[2]: Y origin = 0
byte[2]: image width (little endian)
byte[2]: image height (little endian)
byte: Pixel depth = 32 (=0x20)
byte: Image Descriptor = 0x28 (0x20=image starts from top left, 0x08=8bit for alpha chanel)
Example (where "XX XX" is a width and "YY YY" a height) :
00 00 02 00 00 00 00 00 00 00 00 00 XX XX YY YY 20 28

Data :
One pixel is defined by 4 bytes :
BB GG RR AA
Where :
BB is the blue value (0 or 255)
GG is the green value (0 or 255)
RR is the red value (0 or 255)
AA is the alpha chanel value (set to 0)
There are 4*4 pixels for 1 tileset.
The TGA width is map_width*4.
The TGA height is map_height*4.

Color  	Build state  	Walk state  	Fly state
White  	no build 	no walk 	no fly
Red  	build ok 	no walk 	fly ok
Yellow 	build ok 	no walk 	no fly
Green  	build ok 	walk ok 	no fly
Cyan  	no build 	walk ok 	no fly
Blue  	no build 	walk ok 	fly ok
Magenta no build 	no walk 	fly ok
Black  	build ok 	fly ok  	walk ok
```


1.16 The war3map.wpm file : The Path Map File
---------------------------------------------

You know already what is it, see it at chapter 1.14.

```
Header :
char[4]: file ID = 'MP3W'
int: file version = 0
int: path map width (=map_width*4)
int: path map height (=map_height*4)

Data:
Each byte of the data part is a part of a tileset exactly like for the TGA.
Data size: (map_height*4)*(map_with*4) bytes
Flags table:
0x01: 0 (unused)
0x02: 1=no walk, 0=walk ok
0x04: 1=no fly, 0=fly ok
0x08: 1=no build, 0=build ok
0x10: 0 (unused)
0x20: 1=blight, 0=normal
0x40: 1=no water, 0=water
0x80: 1=unknown, 0=normal
```


1.17 The war3map.doo file : The doodad file for trees
-----------------------------------------------------

The file contains the trees definitions and positions. There are 2 differants file types.

Here is the format :

```
Header :
char[4]: file ID = "W3do"
int: file version = 7
int: subversion? (usually set to [09 00 00 00]h, rarely [07 00 00 00]h)
int: number of trees defined
Data :
Each tree is defined by a block of 42 bytes organized like this:
char[4]: Tree ID (can be found in the file "Units\DestructableData.slk")
int: Variation (little endian)
float: Tree X coordinate on the map
float: Tree Y coordinate on the map
float: Tree Z coordinate on the map
float: Tree angle (radian angle value)(degree = radian*180/pi)
float: Tree X scale
float: Tree Y scale
float: Tree Z scale
byte: Tree flags*
byte: Tree life (integer stored in %, 100% is 0x64, 170% is 0xAA for example)
int: Tree ID number in the World Editor (little endian) (each tree has a different one)

*flags :
0= invisible and non-solid tree
1= visible but non-solid tree
2= normal tree (visible and solid)
```

To sum up how it looks :

```
tt tt tt tt vv vv vv vv xx xx xx xx yy yy yy yy zz zz zz zz aa aa aa aa xs xs xs xs ys ys ys ys zs
zs zs zs ff ll dd dd dd dd
where :
tt : type
vv : variation
xx : x coordinate
yy : y coordinate
zz : z coordinates
aa : rotation angle
xs : x scale
ys : y scale
zs : z scale
ff : flags
ll : life
dd : doodad number in the editor
```

Example (this is the second tree of Legend) :
```
4C 54 6C 74 08 00 00 00 00 00 74 45 00 00 70 44 00 10 24 44 E5 CB 96 40 98 85 98 3F 98 85 98 3F 98 85 98 3F 02 64 8D 01 00 00

4C 54 6C 74 --> LTlt (tree type)
08 00 00 00 --> 00000008 = variation #8 (changes the shape of the tree)
00 00 74 45 --> X=3904.
00 00 70 44 --> Y=960.
00 10 24 44 --> Z=656.25
E5 CB 96 40 --> Angle (float value=4.7123895, angle=270°)
98 85 98 3F --> X\_Scale=1.191577
98 85 98 3F --> Y\_Scale=1.191577
98 85 98 3F --> Z\_Scale=1.191577
02 --> tree is solid and selectable
64 --> life=100% of default tree life
8D 01 00 00 --> 0000018D=397, tree number 397
```


After the last tree definition, there we have the special doodads (which can't be edited once they are placed)
```
int: special doodad format version set to '0'
int: number "s" of "special" doodads ("special" like cliffs,...)
Then "s" times a special doodad structure (16 bytes each):
char[4]: doodad ID
int: Z? (0)
int: X? (w3e coordinates)
int: Y? (w3e coordinates)
```


Frozen Throne expansion pack beta format :

```
Header :
char[4]: file ID = "W3do"
int: file version = 8
int: subversion? ([0B 00 00 00]h)
int: number of trees defined
Data :
Each tree is defined by a block of (usually) 50 bytes but in this version the length can vary because of the random item sets. The data is organized like this:
char[4]: Tree ID (can be found in the file "Units\DestructableData.slk")
int: Variation (little endian)
float: Tree X coordinate on the map
float: Tree Y coordinate on the map
float: Tree Z coordinate on the map
float: Tree angle (radian angle value)(degree = radian*180/pi)
float: Tree X scale
float: Tree Y scale
float: Tree Z scale
byte: Tree flags*
byte: Tree life (integer stored in %, 100% is 0x64, 170% is 0xAA for example)
int: Random item table pointer
if -1 -> no item table
if >= 0 -> items from the item table with this number (defined in the w3i) are dropped on death
int: number "n" of item sets dropped on death (this can only be greater than 0 if the item table pointer was -1)
then there is n times a item set structure
int: Tree ID number in the World Editor (little endian) (each tree has a different one)

*flags:
0= invisible and non-solid tree
1= visible but non-solid tree
2= normal tree (visible and solid)
```

To sum up how it looks:
```
tt tt tt tt vv vv vv vv xx xx xx xx yy yy yy yy zz zz zz zz aa aa aa aa xs xs xs xs ys ys ys ys zs
zs zs zs ff ll bb bb bb bb cc cc cc cc dd dd dd dd
where:
tt : type
vv : variation
xx : x coordinate
yy : y coordinate
zz : z coordinates
aa : rotation angle
xs : x scale
ys : y scale
zs : z scale
ff : flags
ll : life
bb : unknown
cc : unknown
dd : doodad number in the editor
```

After the last tree definition, there we have the special doodads (which can't be edited once they are placed)
```
int: special doodad format version set to '0'
int: number "s" of "special" doodads ("special" like cliffs,...)
Then "s" times a special doodad structure (16 bytes each):
char[4]: doodad ID
int: Z? (0)
int: X? (w3e coordinates)
int: Y? (w3e coordinates)
```


1.18 The war3mapUnits.doo file : The Unit and the Item File
-----------------------------------------------------------

The file contains the definitions and positions of all placed units and items of the map.

Here is the format :

```
Header :
char[4]: file ID = "W3do"
int: file version = 7
int: subversion? (often set to [09 00 00 00]h)
int: number of units and items defined
Data :
Each unit/item is defined by a block of bytes (variable length) organized like this:
char[4]: type ID (iDNR = random item, uDNR = random unit)
int: variation
float: coordinate X
float: coordinate Y
float: coordinate Z
float: rotation angle
float: scale X
float: scale Y
float: scale Z
byte: flags*
int: player number (owner) (player1 = 0, 16=neutral passive)
byte: unknown (0)
byte: unknown (0)
int: hit points (-1 = use default)
int: mana points (-1 = use default, 0 = unit doesn't have mana)
int: number "s" of dropped item sets
then we have s times a dropped item sets structures (see below)
int: gold amount (default = 12500)
float: target acquisition (-1 = normal, -2 = camp)
int: hero level (set to1 for non hero units and items)
int: number "n" of items in the inventory
then there is n times a inventory item structure (see below)
int: number "n" of modified abilities for this unit
then there is n times a ability modification structure (see below)
int: random unit/item flag "r" (for uDNR units and iDNR items)
0 = Any neutral passive building/item, in this case we have
  byte[3]: level of the random unit/item,-1 = any (this is actually interpreted as a 24-bit number)
  byte: item class of the random item, 0 = any, 1 = permanent ... (this is 0 for units)
  r is also 0 for non random units/items so we have these 4 bytes anyway (even if the id wasn't uDNR or iDNR)
1 = random unit from random group (defined in the w3i), in this case we have
  int: unit group number (which group from the global table)
  int: position number (which column of this group)
  the column should of course have the item flag set (in the w3i) if this is a random item
2 = random unit from custom table, in this case we have
  int: number "n" of different available units
  then we have n times a random unit structure

int: custom color (-1 = none, 0 = red, 1=blue,...)
int: Waygate: active destination number (-1 = deactivated, else it's the creation number of the target rect as in war3map.w3r)
int: creation number
*flags: may be similar to the war3map.doo flags

Dropped item set format
int: number "d" of dropable items
"d" times dropable items structures:
char[4]: item ID ([00 00 00 00]h = none)
this can also be a random item id (see below)
int: % chance to be dropped

Inventory item format
int: inventory slot (this is the actual slot - 1, so 1 => 0)
char[4]: item id (as in ItemData.slk) 0x00000000 = none
this can also be a random item id (see below)

Ability modification format
char[4]: ability id (as in AbilityData.slk)
int: active for autocast abilities, 0 = no, 1 = active
int: level for hero abilities

Random unit format
char[4]: unit id (as in UnitUI.slk)
this can also be a random unit id (see below)
int: percentual chance of choice

Random item ids
random item ids are of the type char[4] where the 1st letter is "Y" and the 3rd letter is "I"
the 2nd letter narrows it down to items of a certain item types
"Y" = any type
"i" to "o" = item of this type, the letters are in order of the item types in the dropdown box ("i" = charged)
the 4th letter narrows it down to items of a certain level
"/" = any level (ASCII 47)
"0" ... = specific level (this is ASCII 48 + level, so level 10 will be ":" and level 15 will be "?" and so on)

Random unit ids
random unit ids are of the type char[4] where the 1st three letters are "YYU"
the 4th letter narrows it down to units of a certain level
"/" = any level (ASCII 47)
"0" ... = specific level (this is ASCII 48 + level, so level 10 will be ":" and level 15 will be "?" and so on)
```

Frozen Throne expansion pack beta format :

```
Header:
char[4]: file ID = "W3do"
int: file version = 8
int: subversion? (often set to [0B 00 00 00]h)
int: number of units and items defined
Data:
Each unit/item is defined by a block of bytes (variable length) organized like this:
char[4]: type ID (iDNR = random item, uDNR = random unit)
int: variation
float: coordinate X
float: coordinate Y
float: coordinate Z
float: rotation angle
float: scale X
float: scale Y
float: scale Z
byte: flags*
int: player number (owner) (player1 = 0, 16=neutral passive)
byte: unknown (0)
byte: unknown (0)
int: hit points (-1 = use default)
int: mana points (-1 = use default, 0 = unit doesn't have mana)
int: map item table pointer (for dropped items on death)
if -1 => no item table used
if >= 0 => the item table with this number will be dropped on death
int: number "s" of dropped item sets (can only be greater 0 if the item table pointer was -1)
then we have s times a dropped item sets structures (see below)
int: gold amount (default = 12500)
float: target acquisition (-1 = normal, -2 = camp)
int: hero level (set to1 for non hero units and items)
int: strength of the hero (0 = use default)
int: agility of the hero (0 = use default)
int: intelligence of the hero (0 = use default)
int: number "n" of items in the inventory
then there is n times a inventory item structure (see below)
int: number "n" of modified abilities for this unit
then there is n times a ability modification structure (see below)
int: random unit/item flag "r" (for uDNR units and iDNR items)
0 = Any neutral passive building/item, in this case we have
  byte[3]: level of the random unit/item,-1 = any (this is actually interpreted as a 24-bit number)
  byte: item class of the random item, 0 = any, 1 = permanent ... (this is 0 for units)
  r is also 0 for non random units/items so we have these 4 bytes anyway (even if the id wasn't uDNR or iDNR)
1 = random unit from random group (defined in the w3i), in this case we have
  int: unit group number (which group from the global table)
  int: position number (which column of this group)
  the column should of course have the item flag set (in the w3i) if this is a random item
2 = random unit from custom table, in this case we have
  int: number "n" of different available units
  then we have n times a random unit structure

int: custom color (-1 = none, 0 = red, 1=blue,...)
int: Waygate: active destination number (-1 = deactivated, else it's the creation number of the target rect as in war3map.w3r)
int: creation number
*flags: may be similar to the war3map.doo flags

Dropped item set format
int: number "d" of dropable items
"d" times dropable items structures:
char[4]: item ID ([00 00 00 00]h = none)
this can also be a random item id (see below)
int: % chance to be dropped

Inventory item format
int: inventory slot (this is the actual slot - 1, so 1 => 0)
char[4]: item id (as in ItemData.slk) 0x00000000 = none
this can also be a random item id (see below)

Ability modification format
char[4]: ability id (as in AbilityData.slk)
int: active for autocast abilities, 0 = no, 1 = active
int: level for hero abilities

Random unit format
char[4]: unit id (as in UnitUI.slk)
this can also be a random unit id (see below)
int: percentual chance of choice

Random item ids
random item ids are of the type char[4] where the 1st letter is "Y" and the 3rd letter is "I"
the 2nd letter narrows it down to items of a certain item types
"Y" = any type
"i" to "o" = item of this type, the letters are in order of the item types in the dropdown box ("i" = charged)
the 4th letter narrows it down to items of a certain level
"/" = any level (ASCII 47)
"0" ... = specific level (this is ASCII 48 + level, so level 10 will be ":" and level 15 will be "?" and so on)

Random unit ids
random unit ids are of the type char[4] where the 1st three letters are "YYU"
the 4th letter narrows it down to units of a certain level
"/" = any level (ASCII 47)
"0" ... = specific level (this is ASCII 48 + level, so level 10 will be ":" and level 15 will be "?" and so on)
```

1.19 The war3map.w3i file : The Info File
-----------------------------------------

It contains some of the info displayed when you start a game.

Format:
```
int: file format version = 18
int: number of saves (map version)
int: editor version (little endian)
String: map name
String: map author
String: map description
String: players recommended
float[8]: "Camera Bounds" as defined in the JASS file
int[4]: camera bounds complements* (see note 1) (ints A, B, C and D)
int: map playable area width E* (see note 1)
int: map playable area height F* (see note 1)
   *note 1:
   map width = A + E + B
   map height = C + F + D
int: flags
   0x0001: 1=hide minimap in preview screens
   0x0002: 1=modify ally priorities
   0x0004: 1=melee map
   0x0008: 1=playable map size was large and has never been reduced to medium (?)
   0x0010: 1=masked area are partially visible
   0x0020: 1=fixed player setting for custom forces
   0x0040: 1=use custom forces
   0x0080: 1=use custom techtree
   0x0100: 1=use custom abilities
   0x0200: 1=use custom upgrades
   0x0400: 1=map properties menu opened at least once since map creation (?)
   0x0800: 1=show water waves on cliff shores
   0x1000: 1=show water waves on rolling shores
char: map main ground type
Example: 'A'= Ashenvale, 'X' = City Dalaran
int: Campaign background number (-1 = none)
String: Map loading screen text
String: Map loading screen title
String: Map loading screen subtitle
int: Map loading screen number (-1 = none)
String: Prologue screen text
String: Prologue screen title
String: Prologue screen subtitle
int: max number "MAXPL" of players
array of structures: then, there is MAXPL times a player data like described below.
int: max number "MAXFC" of forces
array of structures: then, there is MAXFC times a force data like described below.
int: number "UCOUNT" of upgrade availability changes
array of structures: then, there is UCOUNT times a upgrade availability change like described below.
int: number "TCOUNT" of tech availability changes (units, items, abilities)
array of structures: then, there is TCOUNT times a tech availability change like described below
int: number "UTCOUNT" of random unit tables
array of structures: then, there is UTCOUNT times a unit table like described below

Players data format:
int: internal player number
int: player type
   1=Human, 2=Computer, 3=Neutral, 4=Rescuable
int: player race
   1=Human, 2=Orc, 3=Undead, 4=Night Elf
int: 00000001 = fixed start position
String: Player name
float: Starting coordinate X
float: Starting coordinate Y
int: ally low priorities flags (bit "x"=1 --> set for player "x")
int: ally high priorities flags (bit "x"=1 --> set for player "x")

Forces data format:
int: Foces Flags
0x00000001: allied (force 1)
0x00000002: allied victory
0x00000004: share vision
0x00000010: share unit control
0x00000020: share advanced unit control
int: player masks (bit "x"=1 --> player "x" is in this force)
String: Force name

Upgrade availability change format
int: Player Flags (bit "x"=1 if this change applies for player "x")
char[4]: upgrade id (as in UpgradeData.slk)
int: Level of the upgrade for which the availability is changed (this is actually the level - 1, so 1 => 0)
int Availability (0 = unavailable, 1 = available, 2 = researched)

Tech availability change format
int: Player Flags (bit "x"=1 if this change applies for player "x")
char[4]: tech id (this can be an item, unit or ability)
there's no need for an availability value, if a tech-id is in this list, it means that it's not available

Random unit table format
int: Number "n" of random groups
then follows n times the following data (for each group)
int: Group number
string: Group name
int: Number "m" of positions
positions are the table columns where you can enter the unit/item ids, all units in the same line have the same chance, but belong to different "sets" of the random group, called positions here
int[m]: for each positon is specified if it's a unit table (=0), a building table (=1) or an item table (=2)
int: Number "i" of units/items, this is the number of lines in the table, each position can have that many or fewer entries
now there's "i" times the following structure (for each line)
int: Chance of the unit/item (percentage)
char[m * 4]: for each position are the unit/item id's for this line specified
this can also be random unit/item ids (see bottom of war3mapUnits.doo definition)
a unit/item id of 0x00000000 indicates that no unit/item is created
```

Frozen Throne expansion pack format :

```
int: file format version = 25
int: number of saves (map version)
int: editor version (little endian)
String: map name
String: map author
String: map description
String: players recommended
float[8]: "Camera Bounds" as defined in the JASS file
int[4]: camera bounds complements* (see note 1) (ints A, B, C and D)
int: map playable area width E* (see note 1)
int: map playable area height F* (see note 1)
   *note 1:
   map width = A + E + B
   map height = C + F + D
int: flags
   0x0001: 1=hide minimap in preview screens
   0x0002: 1=modify ally priorities
   0x0004: 1=melee map
   0x0008: 1=playable map size was large and has never been reduced to medium (?)
   0x0010: 1=masked area are partially visible
   0x0020: 1=fixed player setting for custom forces
   0x0040: 1=use custom forces
   0x0080: 1=use custom techtree
   0x0100: 1=use custom abilities
   0x0200: 1=use custom upgrades
   0x0400: 1=map properties menu opened at least once since map creation (?)
   0x0800: 1=show water waves on cliff shores
   0x1000: 1=show water waves on rolling shores
   0x2000: 1=unknown
   0x4000: 1=unknown
   0x8000: 1=unknown
char: map main ground type
   Example: 'A'= Ashenvale, 'X' = City Dalaran
int: Loading screen background number which is its index in the preset list (-1 = none or custom imported file)
String: path of custom loading screen model (empty string if none or preset)
String: Map loading screen text
String: Map loading screen title
String: Map loading screen subtitle
int: used game data set (index in the preset list, 0 = standard)
String: Prologue screen path (usually empty)
String: Prologue screen text (usually empty)
String: Prologue screen title (usually empty)
String: Prologue screen subtitle (usually empty)
int: uses terrain fog (0 = not used, greater 0 = index of terrain fog style dropdown box)
float: fog start z height
float: fog end z height
float: fog density
byte: fog red value
byte: fog green value
byte: fog blue value
byte: fog alpha value
int: global weather id (0 = none, else it's set to the 4-letter-id of the desired weather found in TerrainArt\Weather.slk)
String: custom sound environment (set to the desired sound lable)
char: tileset id of the used custom light environment
byte: custom water tinting red value
byte: custom water tinting green value
byte: custom water tinting blue value
byte: custom water tinting alpha value
int: max number "MAXPL" of players
array of structures: then, there is MAXPL times a player data like described below.
int: max number "MAXFC" of forces
array of structures: then, there is MAXFC times a force data like described below.
int: number "UCOUNT" of upgrade availability changes
array of structures: then, there is UCOUNT times a upgrade availability change like described below.
int: number "TCOUNT" of tech availability changes (units, items, abilities)
array of structures: then, there is TCOUNT times a tech availability change like described below
int: number "UTCOUNT" of random unit tables
array of structures: then, there is UTCOUNT times a unit table like described below
int: number "ITCOUNT" of random item tables
array of structures: then, there is ITCOUNT times a item table like described below

Players data format:
int: internal player number
int: player type
   1=Human, 2=Computer, 3=Neutral, 4=Rescuable
int: player race
   1=Human, 2=Orc, 3=Undead, 4=Night Elf
int: 00000001 = fixed start position
String: Player name
float: Starting coordinate X
float: Starting coordinate Y
int: ally low priorities flags (bit "x"=1 --> set for player "x")
int: ally high priorities flags (bit "x"=1 --> set for player "x")

Forces data format:
int: Foces Flags
0x00000001: allied (force 1)
0x00000002: allied victory
0x00000004: share vision
0x00000010: share unit control
0x00000020: share advanced unit control
int: player masks (bit "x"=1 --> player "x" is in this force)
String: Force name

Upgrade availability change format
int: Player Flags (bit "x"=1 if this change applies for player "x")
char[4]: upgrade id (as in UpgradeData.slk)
int: Level of the upgrade for which the availability is changed (this is actually the level - 1, so 1 => 0)
int Availability (0 = unavailable, 1 = available, 2 = researched)

Tech availability change format
int: Player Flags (bit "x"=1 if this change applies for player "x")
char[4]: tech id (this can be an item, unit or ability)
there's no need for an availability value, if a tech-id is in this list, it means that it's not available

Random unit table format
int: Number "n" of random groups
then we have n times the following data (for each group)
int: Group number
string: Group name
int: Number "m" of positions
positions are the table columns where you can enter the unit/item ids, all units in the same line have the same chance, but belong to different "sets" of the random group, called positions here
int[m]: for each positon is specified if it's a unit table (=0), a building table (=1) or an item table (=2)
int: Number "i" of units/items, this is the number of lines in the table, each position can have that many or fewer entries
now there's "i" times the following structure (for each line)
int: Chance of the unit/item (percentage)
char[m * 4]: for each position are the unit/item id's for this line specified
this can also be a random unit/item ids (see bottom of war3mapUnits.doo definition)
a unit/item id of 0x00000000 indicates that no unit/item is created

Random item table format
int: Number "n" of random item tables
then we have n times the following data (for each item table)
int: Table number
string: Table name
int: Number "m" of item sets on the current item table
then we have m times the following data (for each item set)
int: Number "i" of items on the current item set
then we have i times the following two values (for each item)
int: Percentual chance
char[4]: Item id (as in ItemData.slk)
this can also be a random item id (see bottom of war3mapUnits.doo definition
```

1.20 The war3map.wts file : The Trigger String Data File
--------------------------------------------------------

Open it with notepad and you'll figure out how it works. Each trigger string is defined by a number (trigger ID) and a value for this number. When Warcraft meets a "TRIGSTR\_***" (where "***" is supposed to be a number), it will look in the trigger string table to find the corresponding string and replace the trigger string by that value. The value for a specific trigger ID is set only once by the first definition encountered for this ID: if you have two times the trigger string 0 defined, only the first one will count. The number following "STRING " must be positive: any negative number will be ignored. If text follows "STRING ", it'll be considered as number 0.

String definition:
It always start with "STRING " followed by the trigger string ID number which is supposed to be different for each trigger string. Then "{" indicates the begining of the string value followed by a string which can contain several lines and "}" that indicates the end of the trigger string definition.

Example:
in the .wts file you have:

```
STRING 0
{
<whatever>
}
```

Then either in the .j, in the .w3i or in one of the object editor files, Warcraft finds a TRIGSTR\_000, it'll look in the table for
trigger string number 0 and it'll find that the value to use is "<whatever>" instead of "TRIGSTR\_000".
If there are more than 999 strings another the reference simply becomes one character longer.


1.21 The war3mapMap.blp file : The Minimap Image
------------------------------------------------

The BLP file contain the JPEG header and the JPEG raw data separated.
BLP stands for "Blip" file which I guess is a "BLIzzard Picture".
There are two types of BLPs:
- JPG-BLPs: use JPG compression
- Paletted BLPs: use palettes and 1 or 2 bytes per pixel depending

The general format of JPG-BLPs:

```
Header:
char[4]: file ID ("BLP1")
int: 0 for JPG-BLP, 1 for Paletted
int: 0x00000008 = has alpha, 0x00000000 = no alpha
int: image width
int: image height
int: flag for alpha channel and team colors (usually 3, 4 or 5), 3 and 4 means color and alpha information for paletted files,
5 means only color information, if >=5 on 'unit' textures, it won't show the team color.
int: always 0x00000001, if 0x00000000 the model that uses this texture will be messy.
int[16]: mipmap offset (offset from the begining of the file)
int[16]: mipmap size (size of mipmaps)

If it's a JPG-BLP we go on with:
int: jpg header size (header size) "h" (usually 0x00000270)
byte[h]: header
followed by 0 bytes until the begining of the jpeg data, we can safely erase these 0 bytes if we fix the mipmap offset specified above
byte[16, mipmap size]: starting from each of the 16 mipmap offset addresses we read 'mipmap size' bytes raw jpeg data till the end of the file, having the header and the mipmap data we can process the picture like ordinary JPG files

If it's a paletted BLP we go here:
byte[4, 255]: the BGRA palette defining 256 colors by their BGRA values, each 1 byte
byte[width x height]: the ColorIndex of each pixel from top left to bottom right, ColorIndex refers to the above defined color palette
byte[width x height]: the AlphaIndex of each pixel on a standard greyscale palette for the alpha channel, where 0 is fully transparent and 255 is opaque,
if the picturetype flag is set to 5, the image doesn't have an alpha channel, so this section will be omitted
```

More detailed blp specs by Magos: http://magos.thejefffiles.com/War3ModelEditor/


1.22 The war3map.mmp file : The Menu Minimap
--------------------------------------------

```
Header:
int: unknown (usually 0, maybe the file format)
int: number of datasets

Data:
The size of a dataset is 16 bytes.
int: icon type
   Icons Types:
      00: gold mine
      01: house
      02: player start (cross)
int: X coordinate of the icon on the map
int: Y coordinate of the icon on the map
   Map Coordinates:
      top left: 10h, 10h
      center: 80h, 80h
      bottom right: F0h, F0h
byte[4]: player icon color
   Player Colors (BB GG RR AA = Blue, Green, Red, Alpha Channel):
      03 03 FF FF : Red
      FF 42 00 FF : Blue
      B9 E6 1C FF : Cyan
      81 00 54 FF : Purple
      00 FC FF FF : Yellow
      0E 8A FE FF : Orange
      00 C0 20 FF : Green
      B0 5B E5 FF : Pink
      97 96 95 FF : Light gray
      F1 BF 7E FF : Light blue
      46 62 10 FF : Aqua
      04 2A 49 FF : Brown
      FF FF FF FF : None
```


1.23 The war3map.w3u file : The Custom Unit File
------------------------------------------------

W3U files have a initial long and then comes two tables. Both look the same.
First table is original units table (Original Blizzard Units).
Second table is user-created units table (Units created by the map designer).

```
Header:
int: W3U Version = 1
x bytes: Original Units table*
y bytes: User-created units table*

Data:
*Table definition:
int: number n of units on this table.
If 0 on original table, then skip default unit table. This is the number of following units. Even if we don't have any changes on original table, this value must be there.
n times a unit definition structure*.

*Unit definition structure:
char[4]: original unit ID (get the IDs from "Units\UnitData.slk" of war3.mpq)
char[4]: new unit ID. If it is on original table, this is 0, since it isn't used.
int: number m of modifications for this unit
m times a modification structure*

*Modification structure:
char[4] modification ID code (get the IDs from "Units\UnitMetaData.slk" of war3.mpq)
int: variable type* t (0=int, 1=real, 2=unreal, 3=String,...)
t type: value (length depends on the type t specified before)
int: end of unit definition (usually 0)

*Variable types:
0=int
1=real
2=unreal
3=string
4=bool
5=char
6=unitList
7=itemList
8=regenType
9=attackType
10=weaponType
11=targetType
12=moveType
13=defenseType
14=pathingTexture
15=upgradeList
16=stringList
17=abilityList
18=heroAbilityList
19=missileArt
20=attributeType
21=attackBits
```

Frozen Throne expansion pack format of "war3map.w3u / w3t / w3b / w3d / w3a / w3h / w3q" : The object data files :

These are the files that store the changes you make in the object editor.
They all have one format in common. They have an initial int and then there are 2 tables
Both look the same. The first table is the original table (Standard Objects by Blizzard).
The second table contains the user created objects (Custom units / items / abilities …)

```
Header:
int: File Version = (usually 1)
x bytes: Original objects table*
y bytes: Custom objects table*

Data:
*Table definition:
int: number n of objects on this table, if 0 on the original table, then skip the default object table. This is the number of following units. Even if we don't have any changes on original table, this value must be there.
n times a object definition structure*.

*Object definition structure:
char[4]: original object ID (see table at the bottom where you can get the object IDs*)
char[4]: new object ID. (if it is on original table, this is 0, since it isn't used)
int: number m of modifications for this object
m times a modification structure*

*Modification structure:
char[4] modification ID (see the table at the bottom where you can get the mod IDs*)
int: variable type* t
[int: level/variation (this integer is only used by some object files depending on the object type, for example the units file doesn’t use this additional integer, but the ability file does, see the table at the bottom to see which object files use this int*) in the ability and upgrade file this is the level of the ability/upgrade, in the doodads file this is the variation, set to 0 if the object doesn't have more than one level/variation]
[int: data pointer (again this int is only used by those object files that also use the level/variation int, see table*) in reality this is only used in the ability file for values that are originally stored in one of the Data columns in AbilityData.slk, this int tells the game to which of those columns the value resolves (0 = A, 1 = B, 2 = C, 3 = D, 4 = F, 5 = G, 6 = H), for example if the change applies to the column DataA3 the level int will be set to 3 and the data pointer to 0]
int, float or string: value of the modification depending on the variable type specified by t
int: end of modification structure (this is either 0, or equal to the original object ID or equal to the new object ID of the current object, when reading files you can use this to check if the format is correct, when writing a file you should use the new object ID of the current object here)

*Variable types:
Value  	Variable Type  	Value Format
0 	Integer 	int
1 	Real 	float (single precision)
2 	Unreal (0 <= val <= 1) 	float (single Precision)
3 	String 	string (null terminated)

*Object data files:
This table shows where to get the object IDs and the modification IDs and if the files use the 2 additional level and data pointer integers.
Extension  	Object Type  	Object IDs  	 	 	Mod IDs  	 	 	 	 	Uses Optional Ints
w3u 	  	Units 	 	Units\UnitData.slk 	 	Units\UnitMetaData.slk 	 	 	 	No
w3t 	  	Items 	 	Units\ItemData.slk 	 	Units\UnitMetaData.slk (those where useItem = 1) No
w3b 	  	Destructables 	Units\DestructableData.slk 	Units\DestructableMetaData.slk 	 	 	No
w3d 	  	Doodads 	Doodads\Doodads.slk 	 	Doodads\DoodadMetaData.slk 	 	 	Yes
w3a 	  	Abilities 	Units\AbilityData.slk 	 	Units\AbilityMetaData.slk 	 	 	Yes
w3h 	  	Buffs 	 	Units\AbilityBuffData.slk 	Units\AbilityBuffMetaData.slk 	 	 	No
w3q 	  	Upgrades 	Units\UpgradeData.slk 	 	Units\UpgradeMetaData.slk 	 	 	Yes
```

These files can also be found in campaign archives with the exactly same format but named war3campaign.w3u / w3t / w3b / w3d / w3a / w3h / w3q


Frozen Throne expansion pack format of w3o object editor files :

The w3o is a collection of all above mentioned object editor files compiled in one single file. You get such a file if you export all object data in the object editor. It can be selected in the world editor as external data source in the map properties dialog, therefore it has to be in the same folder as the map that should use the file.

Format:
```
int: file version (currently 1)
int: contains unit data file (1 = yes, 0 = no)
if yes, then here follows a complete w3u file (see w3u specifications above)
int: contains item data file (1 = yes, 0 = no)
if yes, then here follows a complete w3t file (see w3t specifications above)
int: contains destructable data file (1 = yes, 0 = no)
if yes, then here follows a complete w3b file (see w3b specifications above)
int: contains doodad data file (1 = yes, 0 = no)
if yes, then here follows a complete w3d file (see w3d specifications above)
int: contains ability data file (1 = yes, 0 = no)
if yes, then here follows a complete w3a file (see w3a specifications above)
int: contains buff data file (1 = yes, 0 = no)
if yes, then here follows a complete w3h file (see w3h specifications above)
int: contains upgrade data file (1 = yes, 0 = no)
if yes, then here follows a complete w3q file (see w3q specifications above)
```
