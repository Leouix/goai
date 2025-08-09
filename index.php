<?php

$var = rand(1, 2);

function checkVar() {
    return 'var stringValue';
}

// @todo вызывать checkVar() только если $var === 1

echo checkVar();
