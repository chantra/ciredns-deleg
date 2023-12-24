
# deleg

## Name

*deleg* - Implements the DELEG record

## Description

*deleg* implements the DELEG record. See [draft-dnsop-deleg](https://github.com/fl1ger/deleg/blob/main/draft-dnsop-deleg.md) for details.

## Syntax

For each Server Block, create a `deleg` block with the list of zones it handles and the responses to add for those delegations:

~~~ txt
deleg zone1 zone2 {
    responses "example.org. 3600 IN TXT some text"
}

deleg zone3 {
  responses "example.org. 3600 IN TXT zone3 text"
}
~~~

NOTE: the `owner` of the `deleg` response will be overridden. This syntax is currently used for its simplcity.

## Examples

In this example, we will serve the root zone and add records for the com and org delegations.

First, get a copy of the root zone:

~~~ bash
dig axfr . @xfr.dns.icann.org > rootzone
~~~

set your Corefile as below

~~~ Corefile
. {
  file rootzone
  deleg com org  {
    responses "example.org. 3600 IN TXT \"this is an example\"" "example.com. 3600 IN TXT this is an another example"
  }
  deleg net {
    responses "example.org.  86400  IN DELEG  1 ns1.example.com. ( ipv4hint=192.0.2.1 ipv6hint=2001:DB8::1 )"
  }
}
~~~

Start `coredns`:

~~~ bash
./coredns -conf Corefile -dns.port 1053
~~~

and test:
~~~ bash
dig @corednsserverip -p 1053 foo.org foo.net +noall +auth
~~~

## Enabling the plugin

Add the `deleg` to the list of CoreDNS plugin, right after `dnssec`:

~~~ diff
diff --git a/plugin.cfg b/plugin.cfg
index 532c3dda5..d400b99f4 100644
--- a/plugin.cfg
+++ b/plugin.cfg
@@ -49,6 +49,7 @@ cache:cache
 rewrite:rewrite
 header:header
 dnssec:dnssec
+deleg:github.com/chantra/coredns-deleg
 autopath:autopath
 minimal:minimal
 template:template
 ~~~

## See Also

[draft-dnsop-deleg](https://github.com/fl1ger/deleg/blob/main/draft-dnsop-deleg.md).
