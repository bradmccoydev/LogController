# LogController

## Purpose

The LogController service picks up log messages from the SQS queue, determines which (if any) service should handle them & then dumps them onto the appropriate queue.
