import 'dart:async';
import 'package:ngdart/angular.dart';
import 'package:ngrouter/ngrouter.dart';
import 'package:markdown/markdown.dart' as md;
import 'package:unpub_web/api/models.dart';
import 'package:unpub_web/app_service.dart';
import 'routes.dart';

String markdownToHtml(String markdown) => md.markdownToHtml(
      markdown,
      extensionSet: md.ExtensionSet.gitHubFlavored,
    );

@Component(
  selector: 'detail',
  templateUrl: 'detail_component.html',
  directives: [routerDirectives, coreDirectives],
  exports: [RoutePaths],
  styles: ['.not-exists { margin-top: 100px }'],
  pipes: [DatePipe],
)
class DetailComponent implements OnInit, OnActivate {
  final AppService appService;
  DetailComponent(this.appService);

  late final WebapiDetailView package;
  String? packageName;
  String? packageVersion;
  int activeTab = 0;
  bool packageExists = false;

  String get readmeHtml =>
      package.readme == null ? '' : markdownToHtml(package.readme!);

  String get changelogHtml =>
      package.changelog == null ? '' : markdownToHtml(package.changelog!);

  String get pubDevLink {
    var url = 'https://pub.dev/packages/$packageName';
    if (packageVersion != null) {
      url += '/versions/$packageVersion';
    }
    return url;
  }

  @override
  Future<Null> ngOnInit() async {
    activeTab = 0;
  }

  @override
  void onActivate(_, RouterState current) async {
    final name = current.parameters['name'];
    final version = current.parameters['version'];

    if (name != null) {
      packageName = name;
      packageVersion = version;
      appService.setLoading(true);
      try {
        package = await appService.fetchPackage(name, version);
        packageExists = true;
      } finally {
        appService.setLoading(false);
      }
    }
  }

  getListUrl(String q) {
    return RoutePaths.list.toUrl(queryParameters: {'q': q});
  }

  getDetailUrl(String name, [String? version]) {
    if (version == null) {
      return RoutePaths.detail.toUrl(parameters: {'name': name});
    } else {
      return RoutePaths.detailVersion
          .toUrl(parameters: {'name': name, 'version': version});
    }
  }
}
