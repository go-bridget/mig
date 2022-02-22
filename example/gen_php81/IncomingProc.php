<?php

/** Incoming stats log, writes only */
class IncomingProc
{
	public function __construct(
		/** Tracking ID */
		public $id ?int = 0;

		/** Property name (human readable, a-z) */
		public $property ?mixed = null;

		/** Property Section ID */
		public $property_section ?int = 0;

		/** Property Item ID */
		public $property_id ?int = 0;

		/** Remote IP from user making request */
		public $remote_ip ?mixed = null;

		/** Timestamp of request */
		public $stamp ?string = "";
	) {}
}
