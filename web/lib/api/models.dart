import 'package:json_annotation/json_annotation.dart';

part 'models.g.dart';

@JsonSerializable()
class ListApi {
  final int count;
  final List<ListApiPackage>? packages;

  const ListApi(this.count, this.packages);

  factory ListApi.fromJson(Map<String, dynamic> map) => _$ListApiFromJson(map);
  Map<String, dynamic> toJson() => _$ListApiToJson(this);
}

@JsonSerializable()
class ListApiPackage {
  final String name;
  final String? description;
  final List<String> tags;
  final String latest;
  final DateTime updatedAt;

  const ListApiPackage(
    this.name,
    this.description,
    this.tags,
    this.latest,
    this.updatedAt,
  );

  factory ListApiPackage.fromJson(Map<String, dynamic> map) =>
      _$ListApiPackageFromJson(map);
  Map<String, dynamic> toJson() => _$ListApiPackageToJson(this);
}

@JsonSerializable()
class DetailViewVersion {
  final String version;
  final DateTime createdAt;

  const DetailViewVersion(this.version, this.createdAt);

  factory DetailViewVersion.fromJson(Map<String, dynamic> map) =>
      _$DetailViewVersionFromJson(map);

  Map<String, dynamic> toJson() => _$DetailViewVersionToJson(this);
}

@JsonSerializable()
class WebapiDetailView {
  final String name;
  final String version;
  final String description;
  final String homepage;
  final List<String> uploaders;
  final DateTime createdAt;
  final String? readme;
  final String? changelog;
  final List<DetailViewVersion> versions;
  final List<String> authors;
  final List<String>? dependencies;
  final List<String> tags;

  const WebapiDetailView(
    this.name,
    this.version,
    this.description,
    this.homepage,
    this.uploaders,
    this.createdAt,
    this.readme,
    this.changelog,
    this.versions,
    this.authors,
    this.dependencies,
    this.tags,
  );

  factory WebapiDetailView.fromJson(Map<String, dynamic> map) =>
      _$WebapiDetailViewFromJson(map);

  Map<String, dynamic> toJson() => _$WebapiDetailViewToJson(this);
}
