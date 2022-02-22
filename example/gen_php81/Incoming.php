<?php

/** Incoming stats log, writes only */
class Incoming
{
	public function __construct(
		/** Tracking ID */
		public $id ?int = 0;

		/** Property name (human readable, a-z) */
		public $property ?string = "";

		/** Property Section ID */
		public $property_section ?int = 0;

		/** Property Item ID */
		public $property_id ?int = 0;

		/** Remote IP from user making request */
		public $remote_ip ?string = "";

		/** Timestamp of request */
		public $stamp ?string = "";
	) {}
}
