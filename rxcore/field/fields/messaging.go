/*
   Copyright 2025 The DIRPX Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package fields

const (
	// MessagingSystem is the field key that identifies the messaging or
	// streaming system (for example, "kafka", "rabbitmq").
	MessagingSystem = "messaging_system"

	// MessagingDestination is the field key that records the destination
	// name, such as a topic, queue, or stream.
	MessagingDestination = "messaging_destination"

	// MessagingDestinationKind is the field key that records the kind of
	// destination (for example, "topic", "queue", "stream").
	MessagingDestinationKind = "messaging_destination_kind"

	// MessagingPartition is the field key that records the partition
	// identifier for partitioned messaging systems.
	MessagingPartition = "messaging_partition"

	// MessagingOffset is the field key that records the message offset in
	// a partitioned stream.
	MessagingOffset = "messaging_offset"

	// MessagingKey is the field key that records the message key used for
	// routing or partitioning, when applicable.
	MessagingKey = "messaging_key"
)
