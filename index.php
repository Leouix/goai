<?php
$var = rand(1, 2);
function checkVar()
{
    return 'var stringValue';
}

if ($var === 1) {
    echo checkVar();
}