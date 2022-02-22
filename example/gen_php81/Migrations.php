<?php

/** Migration log of applied migrations */
class Migrations
{
	public function __construct(
		/** Microservice or project name */
		public $project ?mixed = null;

		/** yyyy-mm-dd-HHMMSS.sql */
		public $filename ?mixed = null;

		/** Statement number from SQL file */
		public $statement_index ?int = 0;

		/** ok or full error message */
		public $status ?string = "";
	) {}
}
