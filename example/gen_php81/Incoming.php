<?php

/** Incoming stats log, writes only */
class Incoming
{
	public function __construct(
		/** Tracking ID */
		public ?int $id = 0,

		/** Property name (human readable, a-z) */
		public ?string $property = "",

		/** Property Section ID */
		public ?int $property_section = 0,

		/** Property Item ID */
		public ?int $property_id = 0,

		/** Remote IP from user making request */
		public ?string $remote_ip = "",

		/** Timestamp of request */
		public ?string $stamp = "",
	) {}
}
