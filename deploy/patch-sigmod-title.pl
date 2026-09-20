#!/usr/bin/perl
use strict;
use warnings;
use bytes;

# SigMod 20250703's two extra-loadout class titles call its variadic phrase
# formatter, which returns "%s" here. The package SHA is checked before this
# script runs. Retarget only these two call instructions to its plain phrase
# lookup, and use its existing localized title with no class placeholder.
my ($file, $arch) = @ARGV;
die "usage: $0 <extension> <x86|x64>\n" unless defined $file && defined $arch;
my %calls = (
    x86 => [
        ['e85fc2d4ff', 'e89fc1d4ff'],
        ['e85519d7ff', 'e89518d7ff'],
    ],
    x64 => [
        ['e84b85d2ff', 'e87b84d2ff'],
        ['e8a806d5ff', 'e8d805d5ff'],
    ],
);
die "unknown architecture $arch\n" unless exists $calls{$arch};

open my $in, '<:raw', $file or die "$file: $!\n";
local $/;
my $body = <$in>;
close $in;

my $replace_once = sub {
    my ($old, $new, $label) = @_;
    my $count = () = $body =~ /\Q$old\E/g;
    die "$file: expected one $label, found $count\n" unless $count == 1;
    die "$file: $label changed size\n" unless length($old) == length($new);
    $body =~ s/\Q$old\E/$new/;
};

$replace_once->("Extra loadout items class\0", "Extra loadout items\0" . ("\0" x 6), 'title key');
for my $call (@{$calls{$arch}}) {
    $replace_once->(pack('H*', $call->[0]), pack('H*', $call->[1]), 'title call');
}

my $temporary = "$file.tmp.$$";
open my $out, '>:raw', $temporary or die "$temporary: $!\n";
print {$out} $body or die "$temporary: $!\n";
close $out or die "$temporary: $!\n";
my $mode = (stat $file)[2] & 07777;
chmod $mode, $temporary or die "$temporary: $!\n";
rename $temporary, $file or die "$file: $!\n";
