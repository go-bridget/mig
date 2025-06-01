<?php

/** Migration log of applied migrations */
class Migrations
{
	public function __construct(
		/** Microservice or project name */
		public string $project = "",

		/** yyyy-mm-dd-HHMMSS.sql */
		public string $filename = "",

		/** Statement number from SQL file */
		public int $statement_index = 0,

		/** ok or full error message */
		public string $status = "",
	) {}
}
