#include <ApplicationServices/ApplicationServices.h>
#include <unistd.h>

//gcc -o mosue mouse.c -Wall -framework ApplicationServices


void MoveMouse(int x, int y) {
    CGEventRef move1 = CGEventCreateMouseEvent(
        NULL, kCGEventMouseMoved,
        CGPointMake(x, y),
        kCGMouseButtonLeft // ignored
    );
    CGEventPost(kCGHIDEventTap, move1);
 
    CFRelease(move1);
}

void ClickMouseDown(int x, int y) {
    // Left button down at 250x250
    CGEventRef click1_down = CGEventCreateMouseEvent(
        NULL, kCGEventLeftMouseDown,
        CGPointMake(x, y),
        kCGMouseButtonLeft
    );
    CGEventPost(kCGHIDEventTap, click1_down);
    CFRelease(click1_down);
}
void ClickMouseUp(int x, int y) {
    // Left button up at 250x250
    CGEventRef click1_up = CGEventCreateMouseEvent(
        NULL, kCGEventLeftMouseUp,
        CGPointMake(x, y),
        kCGMouseButtonLeft
    );
 
    CGEventPost(kCGHIDEventTap, click1_up);
    CFRelease(click1_up);
}

