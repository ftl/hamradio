/*
Package dxcc provides information about DXCC prefixes which are stored in a cty.dat file.
The package also provides functions to download, store and update a cty.dat file.
The default remote location for the cty.dat file is http://www.country-files.com/cty/cty.dat.

File Format Description

First line information, fields are separated by ":"
Column  Length  Description
1       26      Country Name
27      5       CQ Zone
32      5       ITU Zone
37      5       2-letter continent abbreviation
42      9       Latitude in degrees, + for North
51      10      Longitude in degrees, + for West
61      9       Local time offset from GMT
70      6       Primary DXCC Prefix (A "*" preceding this prefix
                indicates that the country is on the DARC WAEDC list,
                and counts in CQ-sponsored contests, but not ARRL-
                sponsored contests).

Following lines contain alias DXCC prefixes (including the primary one),
separated by commas (,). Multiple lines are OK; a line to be continued
should end with comma (,) though it's not required. A semi-colon (;)
terminates the last alias prefix in the list.

If an alias prefix is preceded by "=", this indicates that the prefix
is to be treated as a full callsign, i.e. must be an exact match.

The following special characters can be applied after an alias prefix:
(#)     Override CQ Zone
[#]     Override ITU Zone
<#/#>   Override latitude/longitude
{aa}    Override Continent
~#~     Override local time offset from GMT

For detailed information about the file format see http://www.country-files.com/cty-dat-format/

For detailed information about handling of prefixes and suffixes see http://www.cqwpx.com/rules.htm
*/
package dxcc
